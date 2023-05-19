package protocol

import (
	"sync"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
)

// a worker connection wraps all tcp connections from a client to the server
type WorkerConnection interface {
	ID() string
	NumberOfConnections() int
	GeneralConnection
}

type workerConnection struct {
	id           string
	gid          string
	conns        []GeneralConnection
	lastUsedConn int
	mu           *sync.Mutex
}

func NewWorkerConnection(workerID string, conn GeneralConnection) *workerConnection {
	return &workerConnection{
		id:    workerID,
		gid:   utils.RandomUUID(),
		conns: []GeneralConnection{conn},
		mu:    new(sync.Mutex),
	}
}

func (c *workerConnection) ID() string {
	return c.id
}

func (c *workerConnection) getConn() GeneralConnection {
	c.mu.Lock()
	c.lastUsedConn++
	if c.lastUsedConn >= len(c.conns) {
		c.lastUsedConn = 0
	}
	conn := c.conns[c.lastUsedConn]
	c.mu.Unlock()
	return conn
}

func (c *workerConnection) addWorkerConn(conn GeneralConnection) {
	c.mu.Lock()
	c.conns = append(c.conns, conn)
	c.mu.Unlock()
}

func (c *workerConnection) removeWorkerConn(toDelete GeneralConnection) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, conn := range c.conns {
		if toDelete.ConnGID() == conn.ConnGID() {
			c.conns = append(c.conns[:i], c.conns[i+1:]...)
			return
		}
	}
}

func (c *workerConnection) ConnGID() string {
	return c.gid
}

func (c *workerConnection) NumberOfConnections() int {
	return len(c.conns)
}

func (c *workerConnection) Send(m *proto.Message) error {
	return c.getConn().Send(m)
}

func (c *workerConnection) Respond(m *proto.Message, status proto.Status, payload []byte) error {
	return c.getConn().Respond(m, status, payload)
}

func (c *workerConnection) Request(m *proto.Message) (*proto.Message, error) {
	return c.getConn().Request(m)
}

func (c *workerConnection) Address() string {
	return c.getConn().Address()
}

func (c *workerConnection) Close() error {
	multiErr := errors.NewMultiError()
	for _, conn := range c.conns {
		err := conn.Close()
		if err != nil {
			multiErr.Add(err)
		}
	}
	if multiErr == nil {
		return nil
	}
	return multiErr
}
