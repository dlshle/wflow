package api

import "github.com/dlshle/wflow/internal/protocol"

type tcpServer struct {
	protocol.TCPServer
}

func (s *tcpServer) Start() error {
	return s.TCPServer.Start()
}
