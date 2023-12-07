package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/server/job"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	"github.com/robfig/cron"
)

const (
	maxJobAttempts   = 3
	maxAttempRetries = 3
)

type jobWithSchedule struct {
	job           *proto.Job
	executionTime time.Time
	isLocked      atomic.Bool
	attempt       atomic.Uint32
}

type JobScheduler struct {
	ctx           context.Context
	executor      async.Executor
	logger        logging.Logger
	ticker        time.Ticker
	jobDispatcher func(*proto.Job) error
	jobs          map[string]*jobWithSchedule // job_id -> job
	jobHandler    job.Handler
	mutex         *sync.Mutex
}

func NewJobScheduler(ctx context.Context, executor async.Executor, jobDispatcher func(*proto.Job) error, jobHandler job.Handler) *JobScheduler {
	s := &JobScheduler{
		ctx:           ctx,
		executor:      executor,
		logger:        logging.GlobalLogger.WithPrefix("[JobScheduler]"),
		ticker:        *time.NewTicker(time.Second * 5),
		jobs:          make(map[string]*jobWithSchedule),
		jobDispatcher: jobDispatcher,
		jobHandler:    jobHandler,
		mutex:         new(sync.Mutex),
	}
	executor.Execute(s.start)
	return s
}

func (s *JobScheduler) start() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.mutex.Lock()
			jobs := s.jobs
			s.mutex.Unlock()
			for _, job := range jobs {
				s.executor.Execute(func() {
					s.handleJob(job)
				})
			}
		}
	}
}

// only recurring/scheduled job going to be executed in more than a minute will be scheduled
func (s *JobScheduler) Schedule(job *proto.Job) error {
	if job.JobType != proto.JobType_RECURRING && job.JobType != proto.JobType_SCHEDULED {
		return errors.Error("job type " + job.JobType.String() + " is not supported")
	}
	var nextExecutionTime time.Time
	if job.JobType == proto.JobType_SCHEDULED {
		nextExecutionTime = time.Unix(int64(job.GetScheduledTimeSeconds()), 0)
		if nextExecutionTime.Before(time.Now()) {
			// job is expired
			s.setJobToExpired(job)
			return errors.Error("job is expired, scheduled time is " + nextExecutionTime.String())
		}
	} else {
		// cron
		sched, err := cron.Parse(job.GetCronExpression())
		if err != nil {
			return err
		}
		nextExecutionTime = sched.Next(time.Now())
	}

	scheduledJob := &jobWithSchedule{
		job:           job,
		executionTime: nextExecutionTime,
	}

	s.mutex.Lock()
	s.jobs[job.Id] = scheduledJob
	s.mutex.Unlock()
	return nil
}

func (s *JobScheduler) Cancel(jobID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	job := s.jobs[jobID]
	if job == nil {
		return errors.Error("job " + jobID + " is not scheduled")
	}
	if !job.isLocked.Load() {
		job.isLocked.Store(true)
		delete(s.jobs, jobID)
		return nil
	}
	return errors.Error("job " + jobID + " is being processed")
}

// fire job to client if it's a minute to execution, and then client will execute the job at the right time
func (s *JobScheduler) handleJob(scheduledJob *jobWithSchedule) {
	if scheduledJob.isLocked.Load() || s.jobs[scheduledJob.job.Id] == nil {
		return
	}
	if scheduledJob.attempt.Load() > maxAttempRetries {
		s.logger.Error(s.ctx, "job "+scheduledJob.job.Id+" has been attempted for more than 3 times, deleting")
		s.mutex.Lock()
		delete(s.jobs, scheduledJob.job.Id)
		s.mutex.Unlock()
		return
	}
	job := scheduledJob.job
	timeNow := int32(time.Now().Unix())
	scheduledTimeForJob := job.GetScheduledTimeSeconds()
	timeDelta := timeNow - scheduledTimeForJob
	if timeDelta < -60 {
		// skip till next tick
		return
	}
	if timeDelta > 60 {
		// job is a minute later from its scheduled execution time
		if job.JobType == proto.JobType_SCHEDULED {
			s.setJobToExpired(job)
			s.mutex.Lock()
			delete(s.jobs, job.Id)
			s.mutex.Unlock()
			return
		}
	}
	scheduledJob.isLocked.Store(true)
	defer func() {
		scheduledJob.isLocked.Store(false)
	}()
	persisted, err := s.jobHandler.Get(job.Id)
	if err != nil {
		s.logger.Warn(s.ctx, "Failed to get job from store due to "+err.Error()+" will fire job to worker without checking")
	} else if persisted == nil {
		s.logger.Info(s.ctx, "Job "+job.Id+" is not deleted, skipping")
		s.mutex.Lock()
		delete(s.jobs, job.Id)
		s.mutex.Unlock()
		return
	}
	/*
		TODO: what happens if job is timedout?
	*/
	switch job.JobType {
	case proto.JobType_RECURRING:
		s.handleRecurrdingJob(scheduledJob)
	case proto.JobType_SCHEDULED:
		// for scheduled job(once), if the job is already scheduled and created, that means it's already dispatched to worker
		childJob, err := s.jobHandler.GetLatestJobByParentJobID(job.Id)
		if err != nil {
			s.logger.Warn(s.ctx, "Failed to get latest child job for recurring job "+job.Id+" due to "+err.Error())
		} else if childJob != nil {
			s.logger.Warn(s.ctx, "Scheduled job "+job.Id+" is already scheduled, skipping")
			return
		}
		// in this process, job will be dispatched and persisted(with new job ID) to store
		err = s.dispatchAndPersistJobWithRetry(job)
		scheduledJob.attempt.Add(1)
		if err == nil {
			s.mutex.Lock()
			delete(s.jobs, job.Id)
			s.mutex.Unlock()
		}
	}
}

func (s *JobScheduler) handleRecurrdingJob(jobWithSchedule *jobWithSchedule) {
	job := jobWithSchedule.job
	executionTime := jobWithSchedule.executionTime
	childJob, err := s.jobHandler.GetLatestJobByParentJobID(job.Id)
	if err != nil {
		s.logger.Warn(s.ctx, "Failed to get latest child job for recurring job "+job.Id+" due to "+err.Error())
		return
	}
	if childJob != nil {
		if time.Since(time.Unix(int64(childJob.Job.CreatedAt), 0)) > time.Second*time.Duration(30) {
			s.logger.Warn(s.ctx, "Recurring job "+job.Id+" is already scheduled for this moment, skipping")
			return
		}
	}
	s.logger.Info(s.ctx, "Scheduling recurring job "+job.Id)
	childJobInner := s.createScheduledChildJob(job, executionTime)
	jobWithSchedule.attempt.Add(1)
	// in this process, job will be dispatched and persisted(with new job ID) to store
	err = s.dispatchAndPersistJobWithRetry(childJobInner)
	if err != nil {
		s.logger.Warn(s.ctx, "Failed to dispatch recurring job "+job.Id+" due to "+err.Error())
		return
	}
	sched, _ := cron.Parse(job.GetCronExpression())
	s.mutex.Lock()
	s.jobs[job.Id].executionTime = sched.Next(time.Now())
	s.mutex.Unlock()
}

func (s *JobScheduler) createScheduledChildJob(job *proto.Job, executionTime time.Time) *proto.Job {
	return &proto.Job{
		JobType:     proto.JobType_SCHEDULED,
		ActivityId:  job.ActivityId,
		Param:       job.Param,
		ParentJobId: job.Id,
		JobSchedule: &proto.Job_ScheduledTimeSeconds{
			ScheduledTimeSeconds: int32(executionTime.Unix()),
		},
	}
}

func (s *JobScheduler) persistChildJob(job *proto.Job) error {
	jobReport := &proto.JobReport{
		Job:    job,
		Status: proto.JobStatus_DISPATCHED,
	}
	_, err := s.jobHandler.Put(jobReport)
	return err
}

func (s *JobScheduler) dispatchAndPersistJobWithRetry(job *proto.Job) error {
	return utils.RetryWithoutWait(func() error {
		return s.jobDispatcher(job)
	}, maxAttempRetries)
}

func (s *JobScheduler) setJobToExpired(job *proto.Job) {
	jobReport := &proto.JobReport{
		Job:           job,
		Status:        proto.JobStatus_CANCELLED,
		FailureReason: "job is expired, failde to schedule the job",
	}
	s.jobHandler.Put(jobReport)
}
