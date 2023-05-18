package relationmapping

type JobWorkerMapping struct {
	ID       int    `db:"id"`
	JobID    string `db:"job_id"`
	WorkerID string `db:"worker_id"`
}

type ActivityWorkerMapping struct {
	ID         int    `db:"id"`
	ActivityID string `db:"activity_id"`
	WorkerID   string `db:"worker_id"`
}
