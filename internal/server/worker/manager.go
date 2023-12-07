package worker

import (
	"context"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/utils"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/activity"
	"github.com/dlshle/wflow/internal/server/config"
	"github.com/dlshle/wflow/internal/server/job"
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/internal/server/scheduler"
	"github.com/dlshle/wflow/pkg/store"
	wutils "github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	"github.com/robfig/cron"

	gproto "google.golang.org/protobuf/proto"
)

type Manager interface {
	RegisterTCPServer(server protocol.TCPServer)
	GetWorkerConnection(id string) protocol.WorkerConnection
	GetConnectedWorkers() (workers []protocol.WorkerConnection)
	GetWorkerByIDs(ids []string) (workers []*proto.Worker, err error)
	DisconnectWorker(ctx context.Context, workerID string) error
	HandleWorkerUpdate(ctx context.Context, worker *proto.Worker) error
	QueryWorkerFromDB(ctx context.Context, workerID string) (*proto.Worker, error)
	QueryRemoteWorker(ctx context.Context, workerID string) (*proto.Worker, error)
	GetJob(ctx context.Context, jobID string) (*proto.JobReport, error)
	ScheduleJob(ctx context.Context, job *proto.Job) (*proto.JobReport, error)
	DispatchJob(ctx context.Context, activityID, workerID string, param []byte) (*proto.JobReport, error)
	CancelJob(ctx context.Context, jobID string) error
}

func NewManager(cfg config.ServerConfig, workerStore Store, relationMappingHandler relationmapping.Handler, jobHandler job.Handler, activityHandler activity.Handler) Manager {
	logger := logging.GlobalLogger.WithPrefix("[WorkerManager]")
	ctx := context.Background()
	scheduledJobsExecutor := async.NewAsyncPool("scheduled_jobs_executor", cfg.Scheduler.ExecutorMaxJobSize, cfg.Scheduler.ExecutorPoolSize)
	m := &manager{
		ctx:                    ctx,
		logger:                 logger,
		scheduledJobExecutor:   scheduledJobsExecutor,
		workerStore:            workerStore,
		relationMappingHandler: relationMappingHandler,
		jobHandler:             jobHandler,
		activityHandler:        activityHandler,
	}
	jobScheduler := scheduler.NewJobScheduler(ctx, scheduledJobsExecutor, m.dispatchScheduledJob, jobHandler)
	m.jobScheduler = jobScheduler
	err := m.loadAndInitializeJobsForScheduler()
	if err != nil {
		m.logger.Errorf(ctx, "[NewManager] failed to load and initialize jobs for scheduler due to %s", err.Error())
	}
	return m
}

type manager struct {
	ctx                    context.Context
	logger                 logging.Logger
	scheduledJobExecutor   async.Executor
	workerStore            Store
	relationMappingHandler relationmapping.Handler
	jobHandler             job.Handler
	activityHandler        activity.Handler
	server                 protocol.TCPServer
	scheduledJobChan       chan *proto.Job
	jobScheduler           *scheduler.JobScheduler
}

func (m *manager) loadAndInitializeJobsForScheduler() error {
	recurringJobs, err := m.jobHandler.GetInCompletedJobsByType(proto.JobType_RECURRING)
	if err != nil {
		return err
	}
	m.logger.Infof(m.ctx, "[loadAndInitializeJobsForScheduler] found %d incomplete recurring jobs", len(recurringJobs))
	scheduledJobs, err := m.jobHandler.GetInCompletedJobsByType(proto.JobType_SCHEDULED)
	if err != nil {
		return err
	}
	m.logger.Infof(m.ctx, "[loadAndInitializeJobsForScheduler] found %d incomplete scheduled jobs", len(recurringJobs))
	multiErr := errors.NewMultiError()
	toBeScheduledJobs := append(recurringJobs, scheduledJobs...)
	for _, job := range toBeScheduledJobs {
		err := m.jobScheduler.Schedule(job.Job)
		if err != nil {
			multiErr.Add(err)
		}
	}
	if multiErr.Size() == 0 {
		return nil
	}
	m.logger.Infof(m.ctx, "[loadAndInitializeJobsForScheduler] failed to schedule %d jobs", multiErr.Size())
	return multiErr
}

func (m *manager) RegisterTCPServer(server protocol.TCPServer) {
	m.server = server
	m.server.OnWorkerDisconnected(m.handleWorkerDisconnected)
}

func (m *manager) GetWorkerConnection(id string) protocol.WorkerConnection {
	workerConn := m.server.GetWorkerConnectionByID(id)
	if workerConn == nil {
		return nil
	}
	return workerConn
}

func (m *manager) GetConnectedWorkers() (workers []protocol.WorkerConnection) {
	connectedWorkerIDs := m.server.ConnectedWorkerIDs()
	workers = make([]protocol.WorkerConnection, 0)
	for _, id := range connectedWorkerIDs {
		connectedWorker := m.server.GetWorkerConnectionByID(id)
		workers = append(workers, connectedWorker)
	}
	return
}

func (m *manager) GetWorkerByIDs(ids []string) (workers []*proto.Worker, err error) {
	if len(ids) == 0 {
		return []*proto.Worker{}, nil
	}
	workers = make([]*proto.Worker, 0)
	err = m.workerStore.WithTx(func(tx store.SQLTransactional) error {
		for _, id := range ids {
			worker, err := m.workerStore.TxGet(tx, id)
			if err != nil {
				return err
			}
			workers = append(workers, worker)
		}
		return nil
	})
	return
}

func (m *manager) DisconnectWorker(ctx context.Context, workerID string) error {
	m.logger.Infof(ctx, "[DisconnectWorker] disconnecting worker %s", workerID)
	return m.server.DisconnectWorkerConnections(workerID)
}

func (m *manager) HandleWorkerUpdate(ctx context.Context, worker *proto.Worker) (err error) {
	worker, err = m.workerStore.Put(worker)
	if err != nil {
		return
	}
	return m.workerStore.WithTx(func(tx store.SQLTransactional) error {
		return utils.ProcessWithErrors(func() error {
			for _, activity := range worker.SupportedActivities {
				_, err = m.activityHandler.TxPut(tx, activity)
				if err != nil {
					m.logger.Errorf(ctx, "[HandleWorkerUpdate] failed to put activity %v due to %s", activity, err.Error())
					return err
				}
			}
			return nil
		}, func() error {
			return m.relationMappingHandler.TxUpdateWorkerActivities(ctx, tx, worker)
		}, func() error {
			return m.relationMappingHandler.TxUpdateWorkerJobs(ctx, tx, worker)
		})
	})
}

func (m *manager) QueryRemoteWorker(ctx context.Context, workerID string) (*proto.Worker, error) {
	if ctx == nil {
		ctx = m.ctx
	}
	ctx = logging.WrapCtx(ctx, "workerID", workerID)
	workerConn := m.GetWorkerConnection(workerID)
	if workerConn == nil {
		return nil, errors.Error("worker " + workerID + " is not connected")
	}
	return m.queryRemoteWorker(ctx, workerConn)
}

func (m *manager) queryRemoteWorker(ctx context.Context, workerConn protocol.WorkerConnection) (*proto.Worker, error) {
	workerHolder := &proto.Worker{Id: workerConn.ID()}
	workerQueryRequestData, err := gproto.Marshal(workerHolder)
	if err != nil {
		return nil, err
	}
	resp, err := workerConn.Request(&proto.Message{
		Id:      wutils.RandomUUID(),
		Type:    proto.Type_QUERY_WORKER_STATUS,
		Payload: workerQueryRequestData,
	})
	if err != nil {
		return nil, err
	}
	if resp.Status != proto.Status_OK {
		m.logger.Errorf(ctx, "[QueryRemoteWorker] failed to query worker %s due to %s", workerConn.ID(), resp.Status.String())
		return nil, errors.Error("failed to query worker status, remote respond with: " + string(resp.Payload))
	}
	err = gproto.Unmarshal(resp.Payload, workerHolder)
	return workerHolder, err
}

func (m *manager) QueryWorkerFromDB(ctx context.Context, workerID string) (*proto.Worker, error) {
	return m.workerStore.Get(workerID)
}

func (m *manager) ScheduleJob(ctx context.Context, job *proto.Job) (*proto.JobReport, error) {
	err := validateJobRequest(job)
	if err != nil {
		return nil, err
	}
	job.ParentJobId = ""
	// for scheduled job(once), its parent is itself
	if job.JobType == proto.JobType_SCHEDULED {
		job.ParentJobId = job.Id
	}
	jobReport := &proto.JobReport{
		Job:      job,
		WorkerId: "",
		Status:   proto.JobStatus_DISPATCHED,
	}
	jobReport, err = m.jobHandler.Put(jobReport)
	if err != nil {
		return nil, err
	}
	err = m.jobScheduler.Schedule(job)
	return jobReport, err
}

func validateJobRequest(job *proto.Job) error {
	if job.JobType == proto.JobType_RECURRING {
		cronExpression := job.GetCronExpression()
		if cronExpression == "" {
			return errors.Error("cron expression is required for recurring job")
		}
		_, err := cron.Parse(cronExpression)
		return err
	}
	if job.JobType == proto.JobType_SCHEDULED {
		scheduleTimeSeconds := job.GetScheduledTimeSeconds()
		scheduleTime := time.Unix(int64(scheduleTimeSeconds), 0)
		if scheduleTime.Before(time.Now()) {
			return errors.Error("scheduled time must be in the future")
		}
		return nil
	}
	return nil
}

func (m *manager) DispatchJob(ctx context.Context, activityID, workerID string, param []byte) (*proto.JobReport, error) {
	ctx = logging.WrapCtx(ctx, "activityID", activityID)
	ctx = logging.WrapCtx(ctx, "workerID", workerID)
	// TODO: need to support job dispatching in offline mode
	var (
		workerConn protocol.WorkerConnection
		err        error
	)
	if err = m.validateActivity(activityID); err != nil {
		return nil, err
	}
	if workerID == "" {
		workerConn, err = m.findWorkerConnForActivity(activityID)
		if err != nil {
			return nil, err
		}
	} else {
		workerConn = m.GetWorkerConnection(workerID)
	}
	if workerConn == nil {
		return nil, errors.Error("worker " + workerID + " is not connected")
	}
	job := &proto.Job{
		ActivityId: activityID,
		Param:      param,
	}
	jobReport := &proto.JobReport{
		Job:      job,
		WorkerId: workerID,
		Status:   proto.JobStatus_PENDING,
	}
	jobReport, err = m.jobHandler.Put(jobReport)
	if err != nil {
		return nil, err
	}
	dispatchJobRequestData, err := gproto.Marshal(jobReport.Job)
	if err != nil {
		m.setAndPersistJobResultToFailed(jobReport, err)
		return nil, err
	}
	resp, err := workerConn.Request(&proto.Message{
		Id:      wutils.RandomUUID(),
		Type:    proto.Type_DISPATCH_JOB,
		Payload: dispatchJobRequestData,
	})
	if err != nil {
		m.setAndPersistJobResultToFailed(jobReport, err)
		return nil, err
	}
	if resp.Status != proto.Status_OK {
		m.logger.Errorf(ctx, "[DispatchJob] failed to dispatch job %s due to %s", jobReport.Job.Id, resp.Status.String())
		m.setAndPersistJobResultToFailed(jobReport, err)
		return nil, errors.Error("failed to dispatch job, remote respond with: " + string(resp.Payload))
	}
	return jobReport, err
}

func (m *manager) setAndPersistJobResultToFailed(jobReport *proto.JobReport, err error) *proto.JobReport {
	jobReport.Status = proto.JobStatus_FAILED
	jobReport.FailureReason = err.Error()
	m.jobHandler.Put(jobReport)
	return jobReport
}

func (m *manager) validateActivity(activityID string) error {
	if activity, err := m.activityHandler.Get(activityID); err != nil || activity == nil {
		if activity == nil {
			return errors.Error("activity " + activityID + " is not found")
		}
		if err != nil {
			return errors.Error("failed to get activity " + activityID + " due to " + err.Error())
		}
	}
	return nil
}

func (m *manager) findWorkerConnForActivity(activityID string) (workerConn protocol.WorkerConnection, err error) {
	foundWorkers, err := m.relationMappingHandler.FindWorkersByActivityID(activityID)
	if err != nil {
		return nil, errors.Error("can not find worker by activity " + activityID + " due to " + err.Error())
	}
	for _, worker := range foundWorkers {
		if workerConn = m.server.GetWorkerConnectionByID(worker.Id); workerConn != nil {
			return workerConn, nil
		}
	}
	return nil, errors.Error("can not find active worker for activity " + activityID)
}

func (m *manager) GetJob(ctx context.Context, jobID string) (*proto.JobReport, error) {
	return m.jobHandler.Get(jobID)
}

func (m *manager) CancelJob(ctx context.Context, jobID string) error {
	job, err := m.jobHandler.Get(jobID)
	if err != nil {
		return err
	}
	if job.Status == proto.JobStatus_CANCELLED {
		return errors.Error("job " + jobID + " is already cancelled")
	}
	if job.Status == proto.JobStatus_FAILED || job.Status == proto.JobStatus_SUCCESS {
		return errors.Error("job " + jobID + " is already finished")
	}
	workerConn := m.GetWorkerConnection(job.WorkerId)
	if workerConn == nil {
		return errors.Error("worker " + job.WorkerId + " is not connected")
	}
	cancelJobRequestData, err := gproto.Marshal(&proto.Job{Id: jobID})
	if err != nil {
		return err
	}
	resp, err := workerConn.Request(&proto.Message{
		Id:      wutils.RandomUUID(),
		Type:    proto.Type_CANCEL_JOB,
		Payload: cancelJobRequestData,
	})
	if err != nil {
		return err
	}
	if resp.Status != proto.Status_OK {
		m.logger.Errorf(ctx, "[CancelJob] failed to cancel job %s due to %s", jobID, resp.Status.String())
		return errors.Error("failed to cancel job, remote respond with: " + string(resp.Payload))
	}
	return nil
}

func (m *manager) handleWorkerDisconnected(workerID string) {
	err := m.relationMappingHandler.DeleteWorkerMappings(m.ctx, workerID)
	if err != nil {
		m.logger.Errorf(m.ctx, "[handleWorkerDisconnected] failed to delete worker mappings for worker %s due to %s", workerID, err.Error())
	}
}

func (m *manager) dispatchScheduledJob(job *proto.Job) error {
	var (
		err        error
		workerConn protocol.WorkerConnection
		jobReport  *proto.JobReport
	)
	err = utils.ProcessWithErrors(func() error {
		// check activity
		return m.validateActivity(job.ActivityId)
	}, func() error {
		// find worker
		workerConn, err = m.findWorkerConnForActivity(job.ActivityId)
		return err
	}, func() error {
		// persist job status
		jobReport = &proto.JobReport{
			Job:      job,
			WorkerId: workerConn.ID(),
			Status:   proto.JobStatus_PENDING,
		}
		jobReport, err = m.jobHandler.Put(jobReport)
		return err
	}, func() error {
		// dispatch job
		dispatchJobRequestData, err := gproto.Marshal(jobReport.Job)
		if err != nil {
			m.setAndPersistJobResultToFailed(jobReport, err)
			return err
		}
		resp, err := workerConn.Request(&proto.Message{
			Id:      wutils.RandomUUID(),
			Type:    proto.Type_DISPATCH_JOB,
			Payload: dispatchJobRequestData,
		})
		if err != nil {
			m.setAndPersistJobResultToFailed(jobReport, err)
			return err
		}
		if resp.Status != proto.Status_OK {
			m.setAndPersistJobResultToFailed(jobReport, err)
			return errors.Error("failed to dispatch job, remote respond with: " + string(resp.Payload))
		}
		return err
	})
	if err != nil {
		m.setAndPersistJobResultToFailed(jobReport, err)
	}
	return err
}
