package http

import (
	"context"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/activity"
	"github.com/dlshle/wflow/internal/server/job"
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
)

type adminService struct {
	logger                 logging.Logger
	jobHandler             job.Handler
	activityHandler        activity.Handler
	relationMappingHandler relationmapping.Handler
	workerManager          worker.Manager
}

func (s *adminService) GetWorkerConnection(ctx context.Context, id string) protocol.WorkerConnection {
	return s.workerManager.GetWorkerConnection(id)
}

func (s *adminService) ListAllActiveWorkerConnections(ctx context.Context) []protocol.WorkerConnection {
	return s.workerManager.GetConnectedWorkers()
}

func (s *adminService) DisconnectWorker(ctx context.Context, workerID string) error {
	return s.workerManager.DisconnectWorker(ctx, workerID)
}

func (s *adminService) GetWorker(ctx context.Context, workerID string) (*proto.Worker, error) {
	workerConn := s.workerManager.GetWorkerConnection(workerID)
	if workerConn != nil {
		return s.workerManager.QueryRemoteWorker(ctx, workerID)
	}
	s.logger.Warnf(ctx, "worker %s is not connected, query from db", workerID)
	return s.workerManager.QueryWorkerFromDB(ctx, workerID)
}

func (s *adminService) GetActivity(ctx context.Context, activityID string) (*proto.Activity, error) {
	return s.activityHandler.Get(activityID)
}

func (s *adminService) ListAllActiveActivities(ctx context.Context) ([]*proto.Activity, error) {
	return s.relationMappingHandler.ListAllActiveActivities()
}

func (s *adminService) GetWorkersByActivityID(ctx context.Context, activityID string) ([]*proto.Worker, error) {
	return s.relationMappingHandler.FindWorkersByActivityID(activityID)
}

func (s *adminService) GetJobByID(ctx context.Context, jobID string) (*proto.JobReport, error) {
	return s.jobHandler.Get(jobID)
}

func (s *adminService) CancelJob(ctx context.Context, jobID string) error {
	return s.workerManager.CancelJob(ctx, jobID)
}
