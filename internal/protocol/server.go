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
	DisconnectWorkerConnections(string) error
	GetWorkerConnectionByID(string) WorkerConnection
	RequestWorkers(string, *proto.Message) (*proto.Message, error)
	Close() error
	Wait() error
}

type tcpServer struct {
	ctx                 context.Context // this must hold logging ctx w/ server_id to s.id
	tcpServer           gts.TCPServer
	id                  string
	logger              logging.Logger
	messageHandler      MessageHandler
	connectedWorkers    map[string]*workerConnection // we gotta make sure each worker represents a process and has a unique id, a worker could have multiple connections
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
		connectedWorkers:    make(map[string]*workerConnection),
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

func (s *tcpServer) DisconnectWorkerConnections(workerID string) error {
	conn := s.GetWorkerConnectionByID(workerID)
	return conn.Close()
}

func (s *tcpServer) GetWorkerConnectionByID(workerID string) WorkerConnection {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	return s.connectedWorkers[workerID]
}

func (s *tcpServer) RequestWorkers(workerID string, m *proto.Message) (*proto.Message, error) {
	worker := s.getConnectedWorkers(workerID)
	if worker == nil {
		return nil, errors.Error("worker " + workerID + " is not connected")
	}
	return worker.Request(m)
}

func (s *tcpServer) Close() error {
	err := s.tcpServer.Stop()
	s.rwLock.Lock()
	s.lastErr = err
	for k, cs := range s.connectedWorkers {
		cs.Close()
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
	for _, c := range s.connectedWorkers {
		err := c.Send(m)
		if err != nil {
			multiErr.Add(err)
		}
	}
	if multiErr.Size() == 0 {
		return nil
	}
	return multiErr
}

func (s *tcpServer) getConnectedWorkers(id string) WorkerConnection {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	worker := s.connectedWorkers[id]
	return worker
}

func (s *tcpServer) removeConnectedWorker(id, addr string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	delete(s.connectedWorkers, id)
}

func (s *tcpServer) addConnectionWorker(workerID string, conn GeneralConnection) *workerConnection {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	workerConn := s.connectedWorkers[workerID]
	if conn != nil {
		workerConn.addWorkerConn(conn)
	} else {
		workerConn = NewWorkerConnection(workerID, conn)
	}
	s.connectedWorkers[workerID] = workerConn
	return workerConn
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
	conn := NewGeneralConnection(c, s.notificationEmitter, DefaultRequestTimeoutMS)
	// register worker to server
	workerConn := s.addConnectionWorker(workerResp.Id, conn)
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
