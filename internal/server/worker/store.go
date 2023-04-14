package worker

import (
	"time"

	"github.com/dlshle/wflow/pkg/store"
	wutil "github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type Store interface {
	Get(id string) (*proto.Worker, error)
	TxGet(tx store.SQLTransactional, id string) (*proto.Worker, error)
	Put(worker *proto.Worker) (*proto.Worker, error)
	TxPut(tx store.SQLTransactional, worker *proto.Worker) (*proto.Worker, error)
	TxDelete(tx store.SQLTransactional, id string) error
	WithTx(cb func(tx store.SQLTransactional) error) error
}

type workerStore struct {
	pbEntityStore store.PBEntityStore
}

func NewStore(connStr string) (Store, error) {
	pbEntityStore, err := store.Open(connStr, "workers")
	if err != nil {
		return nil, err
	}
	return &workerStore{
		pbEntityStore: pbEntityStore,
	}, nil
}

func (s *workerStore) Get(id string) (*proto.Worker, error) {
	return s.TxGet(s.pbEntityStore.GetDB(), id)
}

func (s *workerStore) TxGet(tx store.SQLTransactional, id string) (*proto.Worker, error) {
	pbEntity, err := s.pbEntityStore.TxGet(tx, id)
	if err != nil {
		return nil, err
	}
	worker := &proto.Worker{}
	err = gproto.Unmarshal(pbEntity.Payload, worker)
	return worker, err
}

func (s *workerStore) Put(worker *proto.Worker) (*proto.Worker, error) {
	return s.TxPut(s.pbEntityStore.GetDB(), worker)
}

func (s *workerStore) TxPut(tx store.SQLTransactional, worker *proto.Worker) (*proto.Worker, error) {
	workerData, err := gproto.Marshal(worker)
	if err != nil {
		return nil, err
	}
	if worker.Id == "" {
		worker.Id = wutil.RandomUUID()
	}
	updatedPBEntity, err := s.pbEntityStore.TxPut(tx, &store.PBEntity{worker.Id, workerData, time.Now()})
	if err != nil {
		return nil, err
	}
	worker.Id = updatedPBEntity.ID
	return worker, nil
}

func (s *workerStore) TxDelete(tx store.SQLTransactional, id string) error {
	return s.pbEntityStore.TxDelete(tx, id)
}

func (s *workerStore) WithTx(cb func(tx store.SQLTransactional) error) error {
	return s.pbEntityStore.WithTx(cb)
}
