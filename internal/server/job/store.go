package job

import (
	"time"

	"github.com/dlshle/wflow/pkg/store"
	wutil "github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type Store interface {
	ListJobIDsByActivityID(activityID string) ([]string, error)
	Get(id string) (*proto.JobReport, error)
	TxGet(tx store.SQLTransactional, id string) (*proto.JobReport, error)
	Put(jobReport *proto.JobReport) (*proto.JobReport, error)
	TxPut(tx store.SQLTransactional, jobReport *proto.JobReport) (*proto.JobReport, error)
	TxDelete(tx store.SQLTransactional, id string) error
	WithTx(func(tx store.SQLTransactional) error) error
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

func (s *jobStore) WithTx(cb func(tx store.SQLTransactional) error) error {
	return s.pbEntityStore.WithTx(cb)
}

func (s *jobStore) Put(jobReport *proto.JobReport) (*proto.JobReport, error) {
	return s.TxPut(s.pbEntityStore.GetDB(), jobReport)
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
	isInsert := false
	if jobReport.Job.Id == "" {
		jobReport.Job.Id = wutil.RandomUUID()
		isInsert = true
	}
	jobReportData, err := gproto.Marshal(jobReport)
	if err != nil {
		return nil, err
	}
	updatedPBEntity, err := s.pbEntityStore.TxPut(tx, &store.PBEntity{jobReport.Job.Id, jobReportData, time.Now()})
	if err != nil {
		return nil, err
	}
	jobReport.Job.Id = updatedPBEntity.ID
	if isInsert {
		err = s.pbEntityStore.WithTx(func(tx store.SQLTransactional) error {
			_, err := tx.Exec("UPDATE jobs SET activity_id = $1 WHERE id = $2", jobReport.Job.ActivityId, jobReport.Job.Id)
			return err
		})
	}
	return jobReport, err
}

func (s *jobStore) ListJobIDsByActivityID(activityID string) ([]string, error) {
	var jobIDs []string
	err := s.pbEntityStore.WithTx(func(s store.SQLTransactional) error {
		return s.Select(&jobIDs, "SELECT id FROM jobs WHERE activity_id = $1", activityID)
	})
	return jobIDs, err
}

func (s *jobStore) TxDelete(tx store.SQLTransactional, id string) error {
	return s.pbEntityStore.TxDelete(tx, id)
}
