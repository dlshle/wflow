package job

import (
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type Store interface {
	Get(id string) (*proto.JobReport, error)
	TxGet(tx store.SQLTransactional, id string) (*proto.JobReport, error)
	TxPut(tx store.SQLTransactional, jobReport *proto.JobReport) (*proto.JobReport, error)
	TxDelete(tx store.SQLTransactional, id string) error
}

type jobStore struct {
	pbEntityStore store.PBEntityStore
}

func NewStore(connStr string) (Store, error) {
	pbEntityStore, err := store.Open(connStr, "jobs")
	if err != nil {
		return nil, err
	}
	return &jobStore{
		pbEntityStore: pbEntityStore,
	}, nil
}

func (s *jobStore) Get(id string) (*proto.JobReport, error) {
	return s.TxGet(s.pbEntityStore.GetDB(), id)
}

func (s *jobStore) TxGet(tx store.SQLTransactional, id string) (*proto.JobReport, error) {
	pbEntity, err := s.pbEntityStore.TxGet(tx, id)
	if err != nil {
		return nil, err
	}
	jobReport := &proto.JobReport{}
	err = gproto.Unmarshal(pbEntity.Payload, jobReport)
	return jobReport, err
}

func (s *jobStore) TxPut(tx store.SQLTransactional, jobReport *proto.JobReport) (*proto.JobReport, error) {
	jobReportData, err := gproto.Marshal(jobReport)
	if err != nil {
		return nil, err
	}
	updatedPBEntity, err := s.pbEntityStore.TxPut(tx, &store.PBEntity{jobReport.Job.Id, jobReportData})
	if err != nil {
		return nil, err
	}
	jobReport.Job.Id = updatedPBEntity.ID
	return jobReport, nil
}

func (s *jobStore) TxDelete(tx store.SQLTransactional, id string) error {
	return s.pbEntityStore.TxDelete(tx, id)
}
