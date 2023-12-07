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
	GetInCompletedJobsByType(jobType proto.JobType) (jobs []*proto.JobReport, err error)
	GetLatestJobByParentWorkerID(parentWorkerID string) (*proto.JobReport, error)
	TxGet(tx store.SQLTransactional, id string) (*proto.JobReport, error)
	Put(jobReport *proto.JobReport) (*proto.JobReport, error)
	TxPut(tx store.SQLTransactional, jobReport *proto.JobReport) (*proto.JobReport, error)
	TxDelete(tx store.SQLTransactional, id string) error
	WithTx(func(tx store.SQLTransactional) error) error
}

type jobStore struct {
	pbEntityStore store.PBEntityStore
}

func NewStore(connStr string) (*jobStore, error) {
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
		jobReport.Job.CreatedAt = int32(time.Now().Unix())
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
			_, err := tx.Exec("UPDATE jobs SET activity_id = $1, job_type = $2"+getJobParentIDInsertParam(jobReport.Job)+" WHERE id = $3", jobReport.Job.ActivityId, jobReport.Job.JobType, jobReport.Job.Id)
			return err
		})
	}
	if jobReport.Status == proto.JobStatus_CANCELLED || jobReport.Status == proto.JobStatus_FAILED || jobReport.Status == proto.JobStatus_SUCCESS {
		err = s.pbEntityStore.WithTx(func(tx store.SQLTransactional) error {
			_, err := tx.Exec("UPDATE jobs SET is_completed = true WHERE id = $1", jobReport.Job.Id)
			return err
		})
	}
	return jobReport, err
}

func getJobParentIDInsertParam(job *proto.Job) string {
	if job.ParentJobId == "" {
		return ""
	}
	return " parent_job_id = " + "'" + job.ParentJobId + "'"
}

func (s *jobStore) GetInCompletedJobsByType(jobType proto.JobType) (jobs []*proto.JobReport, err error) {
	err = s.pbEntityStore.WithTx(func(tx store.SQLTransactional) error {
		var payloads [][]byte
		err := tx.Select(&payloads, "SELECT payload FROM jobs WHERE job_type = $1 AND is_completed = false", jobType)
		if err != nil {
			return err
		}
		jobs = make([]*proto.JobReport, len(payloads), len(payloads))
		for i, payload := range payloads {
			jobs[i] = &proto.JobReport{}
			err = gproto.Unmarshal(payload, jobs[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (s *jobStore) GetLatestJobByParentJobID(parentWorkerID string) (*proto.JobReport, error) {
	var job *proto.JobReport
	err := s.pbEntityStore.WithTx(func(tx store.SQLTransactional) error {
		err := tx.Select(&job, "SELECT payload FROM jobs WHERE parent_job_id = $1 ORDER BY created_at DESC LIMIT 1", parentWorkerID)
		return err
	})
	return job, err
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
