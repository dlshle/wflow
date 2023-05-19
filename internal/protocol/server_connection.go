package protocol

// a server connection represents a connection from a client to the server
type ServerConnection interface {
	ID() string
	GeneralConnection
}

type serverConnection struct {
	id string
	GeneralConnection
}

func NewServerConnection(id string, generalConnection GeneralConnection) ServerConnection {
	return &serverConnection{
		id:                id,
		GeneralConnection: generalConnection,
	}
}

func (c *serverConnection) ID() string {
	return c.id
}
