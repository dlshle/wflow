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
	GetWorkerConnection(id string) protocol.WorkerConnection
	HandleWorkerConnectionDisconnected(ctx context.Context, workerID string)
	HandleWorkerConnnection(ctx context.Context, c protocol.WorkerConnection)
	GetConnectedWorkers() (workers []protocol.WorkerConnection)
	DisconnectWorker(ctx context.Context, workerID string) error
	HandleWorkerUpdate(ctx context.Context, worker *proto.Worker) error
	QueryRemoteWorker(ctx context.Context, workerID string) (*proto.Worker, error)
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
		connectedWorkers:       make(map[string]protocol.WorkerConnection),
	}
}

type manager struct {
	ctx                    context.Context
	logger                 logging.Logger
	workerStore            Store
	relationMappingHandler relationmapping.Handler
	jobHandler             job.Handler
	activityHandler        activity.Handler
	rwLock                 *sync.RWMutex
	connectedWorkers       map[string]protocol.WorkerConnection
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

func (m *manager) GetWorkerConnection(id string) (c protocol.WorkerConnection) {
	m.withRead(func() {
		c = m.connectedWorkers[id]
	})
	return
}

func (m *manager) putWorkerConnection(id string, c protocol.WorkerConnection) {
	m.withWrite(func() {
		m.connectedWorkers[id] = c
	})
}

func (m *manager) removeWorkerConnection(id string) {
	m.withWrite(func() {
		delete(m.connectedWorkers, id)
	})
}

func (m *manager) HandleWorkerConnectionDisconnected(ctx context.Context, workerID string) {
	m.logger.Infof(ctx, "[HandleWorkerConnectionLost] worker %s is disconnected", workerID)
	m.removeWorkerConnection(workerID)
}

func (m *manager) HandleWorkerConnnection(ctx context.Context, c protocol.WorkerConnection) {
	m.logger.Infof(ctx, "[HandleWorkerConnnection] worker %s connected", c.ID())
	m.putWorkerConnection(c.ID(), c)
	c.OnDisconnected(func(gc protocol.GeneralConnection, err error) {
		m.HandleWorkerConnectionDisconnected(ctx, c.ID())
		if err != nil {
			m.logger.Warnf(ctx, "worker %s disconnected with error: %s", c.ID(), err.Error())
		}
	})
}

func (m *manager) GetConnectedWorkers() (workers []protocol.WorkerConnection) {
	workers = make([]protocol.WorkerConnection, 0, len(m.connectedWorkers))
	m.withRead(func() {
		for _, worker := range m.connectedWorkers {
			workers = append(workers, worker)
		}
	})
	return
}

func (m *manager) DisconnectWorker(ctx context.Context, workerID string) error {
	m.logger.Infof(ctx, "[DisconnectWorker] disconnecting worker %s", workerID)
	workerConnection := m.GetWorkerConnection(workerID)
	if workerConnection == nil {
		return errors.Error("worker " + workerID + " is not connected")
	}
	return workerConnection.Close()
}

func (m *manager) HandleWorkerUpdate(ctx context.Context, worker *proto.Worker) (err error) {
	if m.GetWorkerConnection(worker.Id) == nil {
		m.logger.Warnf(ctx, "[HandleWorkerUpdate] worker %s is not connected", worker.Id)
	}
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
	workerHolder := &proto.Worker{Id: workerID}
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
		m.logger.Errorf(ctx, "[QueryRemoteWorker] failed to query worker %s due to %s", workerID, resp.Status.String())
		return nil, errors.Error("failed to query worker status, remote respond with: " + string(resp.Payload))
	}
	err = gproto.Unmarshal(resp.Payload, workerHolder)
	return workerHolder, err
}
