package protocol

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

type TCPServer interface {
	Start() error
	StartAsync()
	Broadcast(*proto.Message) error
	ConnectedWorkerIDs() []string
	RequestWorkers(string, *proto.Message) (map[string]*proto.Message, error)
	Close() error
	Wait() error
}

type tcpServer struct {
	ctx                 context.Context // this must hold logging ctx w/ server_id to s.id
	tcpServer           gts.TCPServer
	id                  string
	logger              logging.Logger
	messageHandler      MessageHandler
	connectedWorkers    map[string](map[string]WorkerConnection)
	notificationEmitter notification.WRNotificationEmitter[*proto.Message]
	asyncPool           async.AsyncPool
	startedTime         time.Time
	rwLock              *sync.RWMutex
	lastErr             error
	stopEventWaiter     *async.WaitLock
}

func NewTCPServer(serverID, address string, port int, messageHandler MessageHandler) TCPServer {
	s := &tcpServer{
		ctx:                 logging.WrapCtx(context.Background(), "server_id", serverID),
		id:                  serverID,
		logger:              logging.GlobalLogger.WithPrefix("[TCPServer]"),
		tcpServer:           gts.NewTCPServer(serverID, address, port),
		messageHandler:      messageHandler,
		connectedWorkers:    make(map[string]map[string]WorkerConnection),
		notificationEmitter: notification.New[*proto.Message](DefaultMaxNotificationListeners),
		asyncPool:           async.NewAsyncPool(serverID, DefaultMaxPoolSize, DefaultMaxAsyncPoolWorkerSize),
		rwLock:              new(sync.RWMutex),
	}
	s.init()
	return s
}

func (s *tcpServer) Start() error {
	s.startedTime = time.Now()
	s.stopEventWaiter = async.NewWaitLock()
	s.lastErr = s.tcpServer.Start()
	s.logger.Errorf(s.ctx, "server %s stopped with error: %v", s.id, s.lastErr)
	return s.lastErr
}

func (s *tcpServer) StartAsync() {
	s.asyncPool.Execute(func() {
		s.Start()
	})
}

func (s *tcpServer) ConnectedWorkerIDs() []string {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	res := make([]string, 0)
	for k := range s.connectedWorkers {
		res = append(res, k)
	}
	return res
}

func (s *tcpServer) RequestWorkers(workerID string, m *proto.Message) (map[string]*proto.Message, error) {
	multiErr := errors.NewMultiError()
	workers := s.getConnectedWorkers(workerID)
	if workers == nil {
		return nil, errors.Error("worker " + workerID + " is not connected")
	}
	responses := make(map[string]*proto.Message)
	for _, worker := range workers {
		response, err := worker.Request(m)
		if err != nil {
			multiErr.Add(err)
		} else {
			responses[worker.ConnGID()] = response
		}
	}
	return responses, multiErr
}

func (s *tcpServer) Close() error {
	err := s.tcpServer.Stop()
	s.rwLock.Lock()
	s.lastErr = err
	for k, cs := range s.connectedWorkers {
		for _, c := range cs {
			c.Close()
		}
		delete(s.connectedWorkers, k)
	}
	if s.stopEventWaiter == nil {
		s.stopEventWaiter = async.NewWaitLock()
	}
	s.stopEventWaiter.Open()
	s.rwLock.Unlock()
	return err
}

func (s *tcpServer) Wait() error {
	if s.stopEventWaiter == nil {
		return errors.Error("server isn't started")
	}
	s.stopEventWaiter.Wait()
	return s.lastErr
}

func (s *tcpServer) init() {
	s.tcpServer.OnClientConnected(func(c gts.Connection) {
		ctx := logging.WrapCtx(s.ctx, "address", c.Address())
		s.logger.Infof(ctx, "client %s connected", c.Address())
		// flow to get
		workerConn, err := s.exchangeProtocol(c)
		if err != nil {
			s.logger.Errorf(ctx, "failed to exchange protocol with client %s: %s", c.Address(), err.Error())
			c.Close()
			return
		}
		ctx = logging.WrapCtx(ctx, "worker", workerConn.ID())

		byteMessageHandler := createHandler(s.messageHandler, s.notificationEmitter)
		c.OnMessage(func(b []byte) {
			s.asyncPool.Execute(func() {
				byteMessageHandler(ctx, workerConn, b)
			})
		})
		c.OnError(func(err error) {
			s.logger.Error(ctx, "worker connection encountered error "+err.Error())
			c.Close()
		})
		c.OnClose(func(err error) {
			s.logger.Warn(ctx, "worker connection "+workerConn.ID()+" closed")
			s.removeConnectedWorker(workerConn.ID(), workerConn.ConnGID())
		})
		s.asyncPool.Execute(c.ReadLoop)
	})
}

func (s *tcpServer) Broadcast(m *proto.Message) error {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	multiErr := errors.NewMultiError()
	for _, conns := range s.connectedWorkers {
		for _, c := range conns {
			err := c.Send(m)
			if err != nil {
				multiErr.Add(err)
			}
		}
	}
	return multiErr
}

func (s *tcpServer) getConnectedWorkers(id string) map[string]WorkerConnection {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	workers := s.connectedWorkers[id]
	return workers
}

func (s *tcpServer) removeConnectedWorker(id, addr string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	delete(s.connectedWorkers, id)
}

func (s *tcpServer) addConnectionWorker(workerConn WorkerConnection) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if s.connectedWorkers[workerConn.ID()] == nil {
		s.connectedWorkers[workerConn.ID()] = make(map[string]WorkerConnection)
	}
	s.connectedWorkers[workerConn.ID()][workerConn.ConnGID()] = workerConn
}

func (s *tcpServer) registerWorker(workerConn WorkerConnection) {
	existingConns := s.getConnectedWorkers(workerConn.ID())
	if existingConns != nil {
		if existingConn := existingConns[workerConn.ConnGID()]; existingConn != nil {
			existingConn.Close()
		}
	}
	s.addConnectionWorker(workerConn)
}

func (s *tcpServer) exchangeProtocol(c gts.Connection) (WorkerConnection, error) {
	// client should send the first stream on connected, and then server sends back the server information in message
	s.logger.Infof(s.ctx, "try to read first stream from connected client "+c.Address())
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
	workerConn := NewWorkerConnection(workerInfo.Id, NewGeneralConnection(c, s.notificationEmitter, DefaultRequestTimeoutMS))
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
