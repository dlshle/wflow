package api

import (
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/job"
	"github.com/dlshle/wflow/internal/server/message_processors"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
)

type tcpServer struct {
	protocol.TCPServer
}

func NewTCPServer(serverID, address string, port int, workerManager worker.Manager, jobHandler job.Handler) *tcpServer {
	return &tcpServer{
		TCPServer: protocol.NewTCPServer(serverID, address, port, protocol.NewMessageHandler(map[proto.Type][]protocol.MessageProcessor{
			proto.Type_JOB_UPDATE:   {message_processors.CreateJobUpdateProcessor(workerManager, jobHandler)},
			proto.Type_WORKER_READY: {message_processors.CreateWorkerReadyProcessor(workerManager)},
		})),
	}
}

func (s *tcpServer) Start() error {
	return s.TCPServer.Start()
}
