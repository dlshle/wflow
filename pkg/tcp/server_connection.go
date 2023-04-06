package tcp

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
