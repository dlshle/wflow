package protocol

type WorkerConnection interface {
	GeneralConnection
	ID() string
}

type workerConnection struct {
	id string
	GeneralConnection
}

func NewWorkerConnection(workerID string, conn GeneralConnection) WorkerConnection {
	return &workerConnection{
		id:                workerID,
		GeneralConnection: conn,
	}
}

func (c *workerConnection) ID() string {
	return c.id
}
