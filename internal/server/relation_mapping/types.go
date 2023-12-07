package relationmapping

import "time"

type JobWorkerMapping struct {
	ID        int       `db:"id"`
	JobID     string    `db:"job_id"`
	WorkerID  string    `db:"worker_id"`
	CreatedAt time.Time `db:"created_at"`
}

type ActivityWorkerMapping struct {
	ID         int       `db:"id"`
	ActivityID string    `db:"activity_id"`
	WorkerID   string    `db:"worker_id"`
	CreatedAt  time.Time `db:"created_at"`
}
