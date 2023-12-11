package job

import (
	"context"
	"sync"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/client/activity"
	wlogging "github.com/dlshle/wflow/internal/client/logging"
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
	InitReportingServer(protocol.ServerConnection, *proto.Server) error
}

type jobManager struct {
	ctx              context.Context
	serverConn       protocol.ServerConnection
	serverInfo       *proto.Server
	workerID         string
	logger           logging.Logger
	workerActivities map[string]activity.WorkerActivity
	jobs             map[string]*workerJobReport
	jobPool          async.AsyncPool
	rwLock           *sync.RWMutex
}

func New(workerID string, activityHandlers map[string]activity.WorkerActivity) JobManager {
	return &jobManager{
		ctx:              logging.WrapCtx(context.Background(), "worker_id", workerID),
		workerID:         workerID,
		logger:           logging.GlobalLogger.WithPrefix("[JobManager]"),
		workerActivities: activityHandlers,
		jobs:             make(map[string]*workerJobReport),
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

func (m *jobManager) getJobReportByID(id string) (j *workerJobReport) {
	m.withRead(func() {
		j = m.jobs[id]
	})
	return
}

func (m *jobManager) setCleintJob(id string, jobReport *workerJobReport) {
	m.withWrite(func() {
		m.jobs[id] = jobReport
	})
}

func (m *jobManager) deleteJob(id string) {
	m.withWrite(func() {
		delete(m.jobs, id)
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

func (m *jobManager) ActiveJobIDs() (res []string) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	res = make([]string, 0)
	for _, job := range m.jobs {
		if isJobActive(job.JobReport) {
			res = append(res, job.Job.Id)
		}
	}
	return
}

func (m *jobManager) WorkerInfo() *proto.Worker {
	worker := &proto.Worker{
		Id:                  m.workerID,
		SystemStat:          utils.GetSystemStat(),
		ActiveJobs:          m.ActiveJobIDs(),
		SupportedActivities: m.SupportedActivities(),
		WorkerStatus:        m.getWorkerStatus(),
	}
	if m.serverInfo != nil {
		worker.ConnectedServer = &m.serverInfo.Id
	}
	return worker
}

func (m *jobManager) InitReportingServer(sc protocol.ServerConnection, server *proto.Server) error {
	m.serverConn = sc
	if server != nil {
		m.serverInfo = server
	}
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
	workerActivity := m.workerActivities[job.ActivityId]
	if workerActivity == nil {
		return errors.Error("can not find activity handler for job " + job.Id + " with activity " + job.ActivityId)
	}
	jobReport := m.generateInitalJobReport(job, workerActivity)
	m.setCleintJob(job.Id, jobReport)
	err := m.scheduleJob(job.Id)
	if err != nil {
		return err
	}
	m.reportJobStatus(jobReport) // report dispatched
	return nil
}

func (m *jobManager) generateInitalJobReport(job *proto.Job, workerActivity activity.WorkerActivity) *workerJobReport {
	jobCtx, cancelFunc := context.WithCancel(m.ctx)
	jobCtx = logging.WrapCtx(jobCtx, "job_id", job.Id)
	job.DispatchTimeInSeconds = int32(time.Now().Unix())
	return &workerJobReport{
		jobCtx:     jobCtx,
		cancelFunc: cancelFunc,
		activity:   workerActivity,
		JobReport: &proto.JobReport{
			Job:      job,
			Result:   nil,
			WorkerId: m.workerID,
			Status:   proto.JobStatus_DISPATCHED,
		},
	}
}

func (m *jobManager) scheduleJob(jobID string) error {
	m.jobPool.Execute(func() {
		m.processJob(jobID)
	})
	return nil
}

func (m *jobManager) processJob(jobID string) {
	jobReport := m.getJobReportByID(jobID)
	if jobReport == nil {
		m.logger.Warn(m.ctx, "job "+jobID+" is cancelled before it's being processed")
		return
	}
	jobCtx := jobReport.jobCtx
	remoteLogginCtx, loggingCancelFunc := context.WithCancel(logging.WrapCtx(context.Background(), "job_id", jobID))
	defer loggingCancelFunc()
	jobLogger := logging.GlobalLogger.WithWriter(wlogging.NewWFlowLogWriter(remoteLogginCtx, jobID, m.serverConn))
	m.logger.Info(jobCtx, "process job "+jobID)
	m.checkAndWaitForScheduledJob(jobCtx, jobReport, jobLogger)
	jobLogger.Infof(jobCtx, "job %s started", jobID)
	jobReport.JobStartedTimeSeconds = int32(time.Now().Unix())
	jobReport.Status = proto.JobStatus_RUNNING
	workerActivity := jobReport.activity
	m.reportJobStatus(jobReport) // report started
	result, err := workerActivity.Handler()(jobCtx, jobLogger, jobReport.Job.Param)
	jobLogger.Infof(jobCtx, "job %s completed, err = %v", jobID, err)
	// before statuses are set, check if job is cancelled
	if m.isJobCancelled(jobID) {
		jobLogger.Info(jobCtx, "job is cancelled, ignoring job results and error")
		return
	}
	if err != nil {
		jobLogger.Info(jobCtx, "job failed due to "+err.Error())
		jobReport.Status = proto.JobStatus_FAILED
		jobReport.FailureReason = err.Error()
	} else {
		jobReport.Status = proto.JobStatus_SUCCESS
		jobReport.Result = result
	}
	m.setCleintJob(jobID, jobReport)
	err = m.reportJobStatus(jobReport)
	if err != nil {
		m.logger.Infof(jobCtx, "failed to report job %s status %s due to %s", jobID, jobReport.Status.String(), err.Error())
	}
	// delete this job record after it's completed
	m.deleteJob(jobID)
}

func (m *jobManager) checkAndWaitForScheduledJob(jobCtx context.Context, jobReport *workerJobReport, jobLogger logging.Logger) {
	if jobReport.Job.JobType == proto.JobType_SCHEDULED {
		executionTime := time.Unix(int64(jobReport.Job.GetScheduledTimeSeconds()), 0)
		jobLogger.Infof(jobCtx, "job %s is scheduled to run at %s", jobReport.Job.Id, executionTime.String())
		time.Sleep(time.Until(executionTime))
	}
}

func (m *jobManager) isJobCancelled(jobID string) bool {
	return m.getJobReportByID(jobID).Status == proto.JobStatus_CANCELLED
}

func (m *jobManager) reportJobStatus(jobReport *workerJobReport) error {
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

func isJobActive(job *proto.JobReport) bool {
	return job.Status == proto.JobStatus_DISPATCHED || job.Status == proto.JobStatus_RUNNING || job.Status == proto.JobStatus_PENDING
}
