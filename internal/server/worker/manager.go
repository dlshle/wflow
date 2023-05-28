package worker

import (
	"context"
	"sync"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/utils"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/activity"
	"github.com/dlshle/wflow/internal/server/job"
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/pkg/store"
	wutils "github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"

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
	DispatchJob(ctx context.Context, activityID, workerID string, param []byte) (*proto.JobReport, error)
	CancelJob(ctx context.Context, jobID string) error
}

func NewManager(workerStore Store, relationMappingHandler relationmapping.Handler, jobHandler job.Handler, activityHandler activity.Handler) Manager {
	logger := logging.GlobalLogger.WithPrefix("[WorkerManager]")
	ctx := context.Background()
	return &manager{
		ctx:                    ctx,
		logger:                 logger,
		workerStore:            workerStore,
		relationMappingHandler: relationMappingHandler,
		jobHandler:             jobHandler,
		activityHandler:        activityHandler,
		rwLock:                 &sync.RWMutex{},
	}
}

type manager struct {
	ctx                    context.Context
	logger                 logging.Logger
	workerStore            Store
	relationMappingHandler relationmapping.Handler
	jobHandler             job.Handler
	activityHandler        activity.Handler
	server                 protocol.TCPServer
	rwLock                 *sync.RWMutex
}

func (m *manager) withRead(cb func()) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	cb()
}

func (m *manager) withWrite(cb func()) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	cb()
}

func (m *manager) RegisterTCPServer(server protocol.TCPServer) {
	m.server = server
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

func (m *manager) DispatchJob(ctx context.Context, activityID, workerID string, param []byte) (*proto.JobReport, error) {
	ctx = logging.WrapCtx(ctx, "activityID", activityID)
	ctx = logging.WrapCtx(ctx, "workerID", workerID)
	// TODO: need to support job dispatching in offline mode
	if workerID == "" {
		foundWorkers, err := m.relationMappingHandler.FindWorkersByActivityID(activityID)
		if err != nil {
			return nil, errors.Error("can not find worker by activity " + activityID + " due to " + err.Error())
		}
		for _, worker := range foundWorkers {
			if m.server.GetWorkerConnectionByID(worker.Id) != nil {
				workerID = worker.Id
				break
			}
		}
		if workerID == "" {
			return nil, errors.Error("can not find active worker for activity " + activityID)
		}
	}
	workerConn := m.GetWorkerConnection(workerID)
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
	jobReport, err := m.jobHandler.Put(jobReport)
	if err != nil {
		return nil, err
	}
	dispatchJobRequestData, err := gproto.Marshal(jobReport.Job)
	if err != nil {
		return nil, err
	}
	// TODO should assign job to worker based on capability, but we will assignment the job to the first match for now
	resp, err := workerConn.Request(&proto.Message{
		Id:      wutils.RandomUUID(),
		Type:    proto.Type_DISPATCH_JOB,
		Payload: dispatchJobRequestData,
	})
	if err != nil {
		return nil, err
	}
	if resp.Status != proto.Status_OK {
		m.logger.Errorf(ctx, "[DispatchJob] failed to dispatch job %s due to %s", jobReport.Job.Id, resp.Status.String())
		return nil, errors.Error("failed to dispatch job, remote respond with: " + string(resp.Payload))
	}
	return jobReport, err
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
