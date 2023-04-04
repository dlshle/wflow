package tcp

import (
	"context"
	"sync"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/notification"
	"github.com/dlshle/gts"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type tcpServer struct {
	ctx                 context.Context // this must hold logging ctx w/ server_id to s.id
	id                  string
	logger              logging.Logger
	messageHandler      MessageHandler
	connectedWorkers    map[string]WorkerConnection
	notificationEmitter notification.WRNotificationEmitter[*proto.Message]
	asyncPool           async.AsyncPool
	startedTime         time.Time
	rwLock              *sync.RWMutex
}

func StartTCPServer(workflowServerID string, address string, port int) error {
	var s tcpServer
	server := gts.NewTCPServer(workflowServerID, address, port)
	server.OnClientConnected(func(c gts.Connection) {
		// flow to get
		workerConn, err := s.exchangeProtocol(c)
		if err != nil {
			c.Close()
			return
		}
		ctx := logging.WrapCtx(s.ctx, "address", c.Address())
		ctx = logging.WrapCtx(ctx, "worker", workerConn.ID())

		s.asyncPool.Execute(c.ReadLoop)

		byteMessageHandler := createHandler(s.messageHandler, s.notificationEmitter)
		c.OnMessage(func(b []byte) {
			s.asyncPool.Execute(func() {
				byteMessageHandler(ctx, c, b)
			})
		})
		c.OnError(func(err error) {
			s.logger.Error(ctx, "worker connection encountered error "+err.Error())
		})
		c.OnClose(func(err error) {
			s.removeConnectedWorker(workerConn.ID())
			s.logger.Warn(ctx, "worker connection "+workerConn.ID()+" closed")
		})
	})
	return server.Start()
}

func (s *tcpServer) Broadcast(m *proto.Message) error {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	multiErr := errors.NewMultiError()
	for _, conn := range s.connectedWorkers {
		err := conn.Send(m)
		if err != nil {
			multiErr.Add(err)
		}
	}
	return multiErr
}

func (s *tcpServer) getConnectedWorker(id string) WorkerConnection {
	s.rwLock.RLock()
	defer s.rwLock.Unlock()
	return s.connectedWorkers[id]
}

func (s *tcpServer) removeConnectedWorker(id string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	delete(s.connectedWorkers, id)
}

func (s *tcpServer) addConnectionWorker(workerConn WorkerConnection) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	s.connectedWorkers[workerConn.ID()] = workerConn
}

func (s *tcpServer) registerWorker(workerConn WorkerConnection) {
	existingConn := s.getConnectedWorker(workerConn.ID())
	if existingConn != nil {
		existingConn.Close()
	}
	s.addConnectionWorker(workerConn)
}

func (s *tcpServer) exchangeProtocol(c gts.Connection) (WorkerConnection, error) {
	// client should send the first stream on connected, and then server sends back the server information in message
	firstStream, err := c.Read()
	if err != nil {
		return nil, errors.Error("unable to read worker greeting: " + err.Error())
	}
	workerResp := &proto.Message{}
	err = gproto.Unmarshal(firstStream, workerResp)
	if err != nil {
		return nil, errors.Error("unable to parse worker greeting: " + err.Error())
	}
	if workerResp.Type != proto.Type_PING {
		return nil, errors.Error("unexpected worker greeting message type " + workerResp.Type.String())
	}
	workerInfo := &proto.Worker{}
	err = gproto.Unmarshal(workerResp.Payload, workerInfo)
	if err != nil {
		return nil, err
	}
	err = s.replyServerInformation(c, workerResp)
	if err != nil {
		return nil, err
	}
	workerConn := NewWorkerConnection(workerInfo.Id, NewGeneralConnection(c))
	// register worker to server
	s.registerWorker(workerConn)
	return workerConn, err
}

func (s *tcpServer) replyServerInformation(c gts.Connection, workerMessage *proto.Message) error {
	serverInfo := &proto.Server{
		Id:              s.id,
		UptimeInSeconds: int32(time.Since(s.startedTime).Seconds()),
	}
	serverReply := workerMessage
	serverReply.Type = proto.Type_PONG
	serverInfoData, err := gproto.Marshal(serverInfo)
	if err != nil {
		return err
	}
	serverReply.Payload = serverInfoData
	serverReplyData, err := gproto.Marshal(serverReply)
	if err != nil {
		return err
	}
	return c.Write(serverReplyData)
}
