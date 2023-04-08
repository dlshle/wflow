package relationmapping

import (
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
	"github.com/jmoiron/sqlx"
	gproto "google.golang.org/protobuf/proto"
)

type relationMappingStore struct {
	db *sqlx.DB
}

func (m *relationMappingStore) TxFindJobsByWorkerID(tx store.SQLTransactional, workerID string) ([]*proto.JobReport, error) {
	var pbEntities []store.PBEntity
	err := tx.Select(&pbEntities, "SELECT * FROM jobs WHERE id IN (SELECT job_id FROM job_worker_mappings WHERE worker_id = $1)", workerID)
	if err != nil {
		return nil, err
	}
	jobReports := make([]*proto.JobReport, len(pbEntities), len(pbEntities))
	for i, entity := range pbEntities {
		err = gproto.Unmarshal(entity.Payload, jobReports[i])
		if err != nil {
			return nil, err
		}
	}
	return jobReports, nil
}

func (m *relationMappingStore) TxFindWorkersByActivityID(tx store.SQLTransactional, activityID string) ([]*proto.Worker, error) {
	var pbEntities []store.PBEntity
	err := tx.Select(&pbEntities, "SELECT * FROM workers WHERE id IN (SELECT DISTINCT(worker_id) FROM activity_worker_mappings WHERE activity_id = $1)", activityID)
	if err != nil {
		return nil, err
	}
	workers := make([]*proto.Worker, len(pbEntities), len(pbEntities))
	for i, entity := range pbEntities {
		err = gproto.Unmarshal(entity.Payload, workers[i])
		if err != nil {
			return nil, err
		}
	}
	return workers, nil
}

func (m *relationMappingStore) TxAddActivityWorkerMapping(tx store.SQLTransactional, activityID, workerID string) (exists bool, err error) {
	mapping := &activityWorkerMapping{}
	err = tx.Select(mapping, "SELECT * FROM activity_worker_mappings WHERE activity_id = $1 AND worker_id = $2", activityID, workerID)
	if err != nil {
		return
	}
	if mapping.activityID != "" {
		exists = true
		return
	}
	_, err = tx.Exec("INSERT INTO activity_worker_mappings (activity_id, worker_id) VALUE ($1, $2)", activityID, workerID)
	exists = false
	return
}

func (m *relationMappingStore) TxAddJobWorkerMapping(tx store.SQLTransactional, jobID, workerID string) (exists bool, err error) {
	mapping := &jobWorkerMapping{}
	err = tx.Select(mapping, "SELECT * FROM job_worker_mappings WHERE job_id = $1 AND worker_id = $2", jobID, workerID)
	if err != nil {
		return
	}
	if mapping.jobID != "" {
		exists = true
		return
	}
	_, err = tx.Exec("INSERT INTO job_worker_mappings (job_id, worker_id) VALUE ($1, $2)", jobID, workerID)
	exists = false
	return
}

func (m *relationMappingStore) TxGetActivityIDsByWorkerID(tx store.SQLTransactional, workerID string) (activityIDs []string, err error) {
	activityIDs = make([]string, 0)
	err = tx.Select(&activityIDs, "SELECT activity_id FROM activity_worker_mappings WHERE worker_id = $1", workerID)
	return
}

func (m *relationMappingStore) TxGetJobIDsByWorkerID(tx store.SQLTransactional, workerID string) (jobIDs []string, err error) {
	jobIDs = make([]string, 0)
	err = tx.Select(&jobIDs, "SELECT job_id FROM job_worker_mappings WHERE worker_id = $1", workerID)
	return
}

func (m *relationMappingStore) TxBulkDeleteJobMappingsByWorkerID(tx store.SQLTransactional, jobIDs []string, workerID string) error {
	_, err := tx.Exec("DELETE FROM job_worker_mappings WHERE worker_id = $1 AND job_id IN "+store.MakeInQueryClause(jobIDs), workerID)
	return err
}

func (m *relationMappingStore) TxBulkDeleteActivityMappingsByWorkerID(tx store.SQLTransactional, activitiesIDs []string, workerID string) error {
	_, err := tx.Exec("DELETE FROM activity_worker_mappings WHERE worker_id = $1 AND activity_id IN "+store.MakeInQueryClause(activitiesIDs), workerID)
	return err
}
