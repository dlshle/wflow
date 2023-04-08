package relationmapping

type jobWorkerMapping struct {
	id       int    `db:"id"`
	jobID    string `db:"job_id"`
	workerID string `db:"worker_id"`
}

type activityWorkerMapping struct {
	id         int    `db:"id"`
	activityID string `db:"activity_id"`
	workerID   string `db:"worker_id"`
}
