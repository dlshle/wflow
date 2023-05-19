package logs

import (
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
	"github.com/jmoiron/sqlx"
	gproto "google.golang.org/protobuf/proto"
)

type Store interface {
	BulkPut([]*proto.JobLog) error
	GetLogsByJobID(string) ([]*proto.JobLog, error)
}

type logsStore struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) Store {
	return &logsStore{db: db}
}

func (s *logsStore) BulkPut(logs []*proto.JobLog) error {
	return store.WithSQLXTx(s.db, func(tx store.SQLTransactional) error {
		for _, log := range logs {
			pbData, err := gproto.Marshal(log)
			if err != nil {
				return err
			}
			_, err = tx.Exec("INSERT INTO job_logs (job_id, timestamp, pb) VALUES ($1, $2, $3)", log.JobId, log.Timestamp, pbData)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *logsStore) GetLogsByJobID(jobID string) ([]*proto.JobLog, error) {
	var logs [][]byte
	err := s.db.Select(&logs, "SELECT pb FROM job_logs WHERE job_id = $1 ORDER BY timestamp DESC", jobID)
	if err != nil {
		return nil, err
	}
	pbLogs := make([]*proto.JobLog, len(logs), len(logs))
	for i, log := range logs {
		pbLog := &proto.JobLog{}
		err = gproto.Unmarshal(log, pbLog)
		if err != nil {
			return nil, err
		}
		pbLogs[i] = pbLog
	}
	return pbLogs, nil
}
