package job

import (
	"context"
	"sync"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

const (
	defaultMaxAsyncPoolSize   = 256
	defaultMaxAsyncWorkerSize = 32
)

type JobManager interface {
	Handle(context.Context, *proto.Job) error
	SupportedActivityIDs() []string
	SupportedActivities() []*proto.Activity
	CancelJob(context.Context, *proto.Job) error
	Jobs() []*proto.Job
	JobIDs() []string
	Job(string) (*proto.JobReport, error)
	IsReady() bool
	WorkerInfo() *proto.Worker
	InitReportingServer(protocol.ServerConnection) error
}

type jobManager struct {
	ctx              context.Context
	serverConn       protocol.ServerConnection
	workerID         string
	logger           logging.Logger
	workerActivities map[string]activity.WorkerActivity
	jobs             map[string]*cancellableJobReport
	jobPool          async.AsyncPool
	rwLock           *sync.RWMutex
}

func New(workerID string, activityHandlers map[string]activity.WorkerActivity) JobManager {
	return &jobManager{
		ctx:              logging.WrapCtx(context.Background(), "worker_id", workerID),
		workerID:         workerID,
		logger:           logging.GlobalLogger.WithPrefix("[JobManager]"),
		workerActivities: activityHandlers,
		jobs:             make(map[string]*cancellableJobReport),
		jobPool:          async.NewAsyncPool("JobManager", defaultMaxAsyncPoolSize, defaultMaxAsyncWorkerSize),
		rwLock:           new(sync.RWMutex),
	}
}

func (m *jobManager) withRead(cb func()) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	cb()
}

func (m *jobManager) withWrite(cb func()) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	cb()
}

func (m *jobManager) getJobReportByID(id string) (j *cancellableJobReport) {
	m.withRead(func() {
		j = m.jobs[id]
	})
	return
}

func (m *jobManager) setCancellableJob(id string, jobReport *cancellableJobReport) {
	m.withWrite(func() {
		m.jobs[id] = jobReport
	})
}

func (m *jobManager) SupportedActivityIDs() []string {
	var activitiesIDs []string
	for k := range m.workerActivities {
		activitiesIDs = append(activitiesIDs, k)
	}
	return activitiesIDs
}

func (m *jobManager) SupportedActivities() []*proto.Activity {
	var activities []*proto.Activity
	for _, activity := range m.workerActivities {
		activities = append(activities, activity.Activity())
	}
	return activities
}

func (m *jobManager) Jobs() []*proto.Job {
	var jobs []*proto.Job
	m.withRead(func() {
		for _, job := range m.jobs {
			jobs = append(jobs, job.Job)
		}
	})
	return jobs
}

func (m *jobManager) JobIDs() []string {
	jobs := m.Jobs()
	jobIDs := make([]string, len(jobs), len(jobs))
	for i, job := range jobs {
		jobIDs[i] = job.Id
	}
	return jobIDs
}

func (m *jobManager) Job(jobID string) (*proto.JobReport, error) {
	cancellableJobReport := m.getJobReportByID(jobID)
	if cancellableJobReport == nil {
		return nil, errors.Error("can not find job " + jobID)
	}
	return cancellableJobReport.JobReport, nil
}

func (m *jobManager) WorkerInfo() *proto.Worker {
	return &proto.Worker{
		Id:                  m.workerID,
		SystemStat:          utils.GetSystemStat(),
		ActiveJobs:          m.JobIDs(),
		SupportedActivities: m.SupportedActivities(),
		WorkerStatus:        m.getWorkerStatus(),
	}
}

func (m *jobManager) InitReportingServer(sc protocol.ServerConnection) error {
	m.serverConn = sc
	return m.reportForWorkerReady()
}

func (m *jobManager) reportForWorkerReady() error {
	worker := m.WorkerInfo()
	workerData, err := gproto.Marshal(worker)
	if err != nil {
		return err
	}
	_, err = m.serverConn.Request(&proto.Message{
		Id:      utils.RandomUUID(),
		Type:    proto.Type_WORKER_READY,
		Payload: workerData,
	})
	return err
}

func (m *jobManager) getWorkerStatus() proto.WorkerStatus {
	if m.IsReady() {
		return proto.WorkerStatus_ACTIVE
	}
	return proto.WorkerStatus_ONLINE
}

func (m *jobManager) IsReady() bool {
	isReady := false
	m.withRead(func() {
		isReady = m.serverConn != nil
	})
	return isReady
}

func (m *jobManager) CancelJob(ctx context.Context, job *proto.Job) (err error) {
	m.withWrite(func() {
		jobReport := m.jobs[job.Id]
		if jobReport == nil {
			err = errors.Error("job " + job.Id + " is not found")
			return
		}
		if jobReport.Status == proto.JobStatus_PENDING {
			delete(m.jobs, job.Id)
			m.logger.Info(jobReport.jobCtx, "job "+job.Id+" is deleted due to cancellation")
			return
		}
		m.logger.Info(jobReport.jobCtx, "job "+job.Id+" is set to cancelled")
		jobReport.Status = proto.JobStatus_CANCELLED
		m.jobs[job.Id] = jobReport
	})
	return
}

func (m *jobManager) Handle(ctx context.Context, job *proto.Job) error {
	m.logger.Info(ctx, "received job "+job.Id)
	maybeJobReport := m.getJobReportByID(job.Id)
	if maybeJobReport != nil {
		return errors.Error("job " + job.Id + " already exists")
	}
	if m.workerActivities[job.ActivityId] == nil {
		return errors.Error("can not find activity handler for job " + job.Id + " with activity " + job.ActivityId)
	}
	m.setCancellableJob(job.Id, m.generateInitalJobReport(job))
	m.scheduleJob(job.Id)
	return nil
}

func (m *jobManager) generateInitalJobReport(job *proto.Job) *cancellableJobReport {
	jobCtx, cancelFunc := context.WithCancel(m.ctx)
	jobCtx = logging.WrapCtx(m.ctx, "job_id", job.Id)
	return &cancellableJobReport{
		jobCtx:     jobCtx,
		cancelFunc: cancelFunc,
		JobReport: &proto.JobReport{
			Job:                   job,
			Result:                nil,
			JobStartedTimeSeconds: int32(time.Now().Unix()),
			WorkerId:              m.workerID,
			Status:                proto.JobStatus_DISPATCHED,
		},
	}
}

func (m *jobManager) scheduleJob(jobID string) {
	m.jobPool.Execute(func() {
		m.processJob(jobID)
	})
}

func (m *jobManager) processJob(jobID string) {
	jobReport := m.getJobReportByID(jobID)
	if jobReport == nil {
		m.logger.Warn(m.ctx, "job "+jobID+" is cancelled before it's being processed")
		return
	}
	jobCtx := jobReport.jobCtx
	m.logger.Info(jobCtx, "process job "+jobID)
	jobReport.Status = proto.JobStatus_RUNNING
	workerActivity := m.workerActivities[jobReport.Job.ActivityId]
	m.reportJobStatus(jobReport) // we can do this in async
	result, err := workerActivity.Handler()(jobCtx, jobReport.Job.Param)
	// before statuses are set, check if job is cancelled
	if m.isJobCancelled(jobID) {
		m.logger.Info(jobCtx, "job is cancelled, ignoring job results and error")
		return
	}
	if err != nil {
		m.logger.Info(jobCtx, "job "+jobID+" failed due to "+err.Error())
		jobReport.Status = proto.JobStatus_FAILED
	} else {
		m.logger.Info(jobCtx, "job "+jobID+" is completed successfully")
		jobReport.Status = proto.JobStatus_SUCCESS
		jobReport.Result = result
	}
	m.setCancellableJob(jobReport.Job.Id, jobReport)
	err = m.reportJobStatus(jobReport)
	if err != nil {
		m.logger.Infof(jobCtx, "failed to report job %s status %s due to %s", jobID, jobReport.Status.String(), err.Error())
	}
}

func (m *jobManager) isJobCancelled(jobID string) bool {
	return m.getJobReportByID(jobID).Status == proto.JobStatus_CANCELLED
}

func (m *jobManager) reportJobStatus(jobReport *cancellableJobReport) error {
	if !m.IsReady() {
		return errors.Error("client is not ready")
	}
	jobReportData, err := gproto.Marshal(jobReport)
	if err != nil {
		return err
	}
	_, err = m.serverConn.Request(&proto.Message{
		Id:      utils.RandomUUID(),
		Type:    proto.Type_JOB_UPDATE,
		Payload: jobReportData,
	})
	return err
}
