package service

import (
	"context"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/server/activity"
	"github.com/dlshle/wflow/internal/server/job"
	"github.com/dlshle/wflow/internal/server/logs"
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
)

type AdminService interface {
	ListAllActiveWorkers(ctx context.Context) ([]*proto.Worker, error)
	DisconnectWorker(ctx context.Context, workerID string) error
	GetWorker(ctx context.Context, workerID string) (*proto.Worker, error)
	GetActivity(ctx context.Context, activityID string) (*proto.Activity, error)
	ListAllActiveActivities(ctx context.Context) ([]*proto.Activity, error)
	GetWorkersByActivityID(ctx context.Context, activityID string) (map[string]*proto.Worker, error)
	ListAllActivitiesWithJobIDs(ctx context.Context) ([]*proto.ActivityWithJobIDs, error)
	GetJobByID(ctx context.Context, jobID string) (*proto.JobReport, error)
	CancelJob(ctx context.Context, jobID string) error
	GetLogsByJobID(ctx context.Context, jobID string) ([]*proto.JobLog, error)
	DispatchJob(ctx context.Context, activityID, workerID string, param []byte) (*proto.JobReport, error)
}

type adminService struct {
	logger                 logging.Logger
	jobHandler             job.Handler
	logStore               logs.Store
	activityHandler        activity.Handler
	relationMappingHandler relationmapping.Handler
	workerManager          worker.Manager
}

func NewAdminService(jobHandler job.Handler, activityHandler activity.Handler, relationMappingHandler relationmapping.Handler, workerManager worker.Manager, logStore logs.Store) AdminService {
	return &adminService{
		logger:                 logging.GlobalLogger.WithPrefix("[AdminService]"),
		jobHandler:             jobHandler,
		activityHandler:        activityHandler,
		relationMappingHandler: relationMappingHandler,
		workerManager:          workerManager,
		logStore:               logStore,
	}
}

func (s *adminService) ListAllActiveWorkers(ctx context.Context) ([]*proto.Worker, error) {
	connectedWorkers := s.workerManager.GetConnectedWorkers()
	workerIDs := make([]string, len(connectedWorkers), len(connectedWorkers))
	for i := range connectedWorkers {
		workerIDs[i] = connectedWorkers[i].ID()
	}
	return s.workerManager.GetWorkerByIDs(workerIDs)
}

func (s *adminService) ListAllActivitiesWithJobIDs(ctx context.Context) ([]*proto.ActivityWithJobIDs, error) {
	allActivities, err := s.activityHandler.GetAll()
	if err != nil {
		return nil, err
	}
	activityWithJobIDs := make([]*proto.ActivityWithJobIDs, len(allActivities), len(allActivities))
	for i, activity := range allActivities {
		jobIDs, err := s.jobHandler.ListJobIDsByActivityID(activity.Id)
		if err != nil {
			return nil, err
		}
		activityWithJobIDs[i] = &proto.ActivityWithJobIDs{Activity: activity, JobIds: jobIDs}
	}
	return activityWithJobIDs, nil
}

func (s *adminService) DisconnectWorker(ctx context.Context, workerID string) error {
	return s.workerManager.DisconnectWorker(ctx, workerID)
}

func (s *adminService) GetWorker(ctx context.Context, workerID string) (*proto.Worker, error) {
	workerConn := s.workerManager.GetWorkerConnection(workerID)
	if workerConn == nil {
		s.logger.Warnf(ctx, "worker %s is not connected, query from db", workerID)
		return s.workerManager.QueryWorkerFromDB(ctx, workerID)
	}
	return s.workerManager.QueryRemoteWorker(ctx, workerID)
}

func (s *adminService) GetActivity(ctx context.Context, activityID string) (*proto.Activity, error) {
	return s.activityHandler.Get(activityID)
}

func (s *adminService) ListAllActiveActivities(ctx context.Context) ([]*proto.Activity, error) {
	return s.relationMappingHandler.ListAllActiveActivities()
}

func (s *adminService) GetWorkersByActivityID(ctx context.Context, activityID string) (map[string]*proto.Worker, error) {
	workers, err := s.relationMappingHandler.FindWorkersByActivityID(activityID)
	if err != nil {
		return nil, err
	}
	workersMap := make(map[string]*proto.Worker)
	for _, worker := range workers {
		workersMap[worker.Id] = worker
	}
	connectedWorkers := s.workerManager.GetConnectedWorkers()
	for _, activeWorkerConn := range connectedWorkers {
		activeWorker := workersMap[activeWorkerConn.ID()]
		if activeWorker != nil {
			x := ""
			activeWorker.ConnectedServer = &x
		}
	}
	return workersMap, nil
}

func (s *adminService) GetJobByID(ctx context.Context, jobID string) (*proto.JobReport, error) {
	return s.jobHandler.Get(jobID)
}

func (s *adminService) CancelJob(ctx context.Context, jobID string) error {
	return s.workerManager.CancelJob(ctx, jobID)
}

func (s *adminService) DispatchJob(ctx context.Context, activityID, workerID string, param []byte) (*proto.JobReport, error) {
	if activityID == "" {
		return nil, errors.Error("activity id is empty")
	}
	return s.workerManager.DispatchJob(ctx, activityID, workerID, param)
}

func (s *adminService) GetLogsByJobID(ctx context.Context, jobID string) ([]*proto.JobLog, error) {
	return s.logStore.GetLogsByJobID(jobID)
}
