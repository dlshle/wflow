package protocol

// a worker connection wraps all tcp connections from a client to the server
type WorkerConnection interface {
	ID() string
	GeneralConnection
}

type workerConnection struct {
	id string
	GeneralConnection
}

func NewWorkerConnection(workerID string, conn GeneralConnection) *workerConnection {
	return &workerConnection{
		id:                workerID,
		GeneralConnection: conn,
	}
}

func (c *workerConnection) ID() string {
	return c.id
}
