package protocol

import (
	"context"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/ctimer"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/notification"
	"github.com/dlshle/gts"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	"github.com/gofrs/uuid"
	gproto "google.golang.org/protobuf/proto"
)

type TCPClient interface {
	Request(*proto.Message) (*proto.Message, error)
	Connect() (ServerConnection, error)
	Close() error
}

type tcpClient struct {
	ctx                 context.Context
	id                  string
	logger              logging.Logger
	tcpClient           gts.TCPClient
	messageHandler      MessageHandler
	asyncPool           async.AsyncPool
	notificationEmitter notification.WRNotificationEmitter[*proto.Message]
	supportedActivities []*proto.Activity
	connectedServer     *proto.Server
	serverConn          ServerConnection
}

func NewTCPClient(id, address string, port int, messageHandler MessageHandler, supportedActivities []*proto.Activity) TCPClient {
	c := &tcpClient{
		ctx:                 logging.WrapCtx(context.Background(), "client_id", id),
		id:                  id,
		logger:              logging.GlobalLogger.WithPrefix("[TCPClient]"),
		messageHandler:      messageHandler,
		notificationEmitter: notification.New[*proto.Message](DefaultMaxNotificationListeners),
		asyncPool:           async.NewAsyncPool(id, DefaultMaxPoolSize, DefaultMaxAsyncPoolWorkerSize),
		supportedActivities: supportedActivities,
	}
	c.init()
	return c
}

func (c *tcpClient) init() {
	c.tcpClient.OnConnectionEstablished(func(conn gts.Connection) {
		c.ctx = logging.WrapCtx(c.ctx, "server_ip", conn.Address())
		err := c.exchangeProtocolAndAttachServerConnection(conn)
		if err != nil {
			c.logger.Errorf(c.ctx, "failed to exchange protocol with %s due to %s", conn.Address(), err.Error())
			conn.Close()
			return
		}
		c.asyncPool.Execute(conn.ReadLoop)
		byteMessageHandler := createHandler(c.messageHandler, c.notificationEmitter)
		conn.OnMessage(func(b []byte) {
			c.asyncPool.Execute(func() {
				byteMessageHandler(c.ctx, c.serverConn, b)
			})
		})
		c.healthCheckRoutine()
	})
	c.tcpClient.OnDisconnected(func(err error) {
		c.logger.Warn(c.ctx, "server connection is closed")
		c.serverConn = nil
	})
	c.tcpClient.OnError(func(err error) {
		c.logger.Error(c.ctx, "server connection encountered error "+err.Error())
	})
}

func (c *tcpClient) Connect() (ServerConnection, error) {
	_, err := c.tcpClient.Connect("")
	if err != nil {
		return nil, err
	}
	return c.serverConn, nil
}

func (c *tcpClient) Request(m *proto.Message) (*proto.Message, error) {
	if c.serverConn == nil {
		return nil, errors.Error("connection to server isn't established or is lost")
	}
	return c.serverConn.Request(m)
}

func (c *tcpClient) Close() error {
	err := c.tcpClient.Disconnect()
	c.serverConn = nil
	c.connectedServer = nil
	return err
}

func (c *tcpClient) exchangeProtocolAndAttachServerConnection(conn gts.Connection) error {
	c.logger.Info(c.ctx, "attempting to exchange protocol with server "+conn.String())
	clientStat := utils.GetSystemStat()
	clientInfo := &proto.Worker{
		Id:                  c.id,
		SystemStat:          clientStat,
		SupportedActivities: c.supportedActivities,
		WorkerStatus:        proto.WorkerStatus_ONLINE,
	}
	clientInfoData, err := gproto.Marshal(clientInfo)
	if err != nil {
		return err
	}
	clientInfoRequest := &proto.Message{
		Id:      c.id,
		Type:    proto.Type_PING,
		Payload: clientInfoData,
	}
	clientInfoRequestData, err := gproto.Marshal(clientInfoRequest)
	if err != nil {
		return err
	}
	err = conn.Write(clientInfoRequestData)
	if err != nil {
		return err
	}
	// waiting for server response with server info
	serverInfoData, err := conn.Read()
	if err != nil {
		return err
	}
	serverInfoResponse := &proto.Message{}
	err = gproto.Unmarshal(serverInfoData, serverInfoResponse)
	if err != nil {
		return err
	}
	if serverInfoResponse.Type != proto.Type_PONG {
		c.logger.Error(c.ctx, "unexpected server response type for protocol exchange "+serverInfoResponse.Type.String())
		return err
	}
	serverInfo := &proto.Server{}
	err = gproto.Unmarshal(serverInfoResponse.Payload, serverInfo)
	if err != nil {
		return err
	}
	c.connectedServer = serverInfo
	c.serverConn = NewServerConnection(serverInfo.Id, NewGeneralConnection(conn, DefaultRequestTimeoutMS))
	return nil
}

func (c *tcpClient) healthCheckRoutine() {
	timer := ctimer.New(time.Second*30, c.healthCheck)
	timer.WithAsyncPool(c.asyncPool)
	timer.Repeat()
}

func (c *tcpClient) healthCheck() {
	if c.serverConn == nil {
		// if server connection lost, skip health check
		return
	}
	if err := c.doHealthCheck(); err != nil {
		c.logger.Error(c.ctx, "health check failed connection to server is lost "+err.Error())
		// we need to close and try to reconnect
		c.serverConn.Close()
		c.serverReconnectingLoop()
	}
}

func (c *tcpClient) doHealthCheck() error {
	mID, err := uuid.NewV4()
	if err != nil {
		return err
	}
	_, err = c.serverConn.Request(&proto.Message{
		Id:   mID.String(),
		Type: proto.Type_PING,
	})
	return err
}

func (c *tcpClient) serverReconnectingLoop() {
	c.logger.Info(c.ctx, "initiating server reconnecting loop")
	for c.serverConn == nil {
		_, err := c.Connect()
		if err == nil {
			c.logger.Info(c.ctx, "server is reconnected")
			return
		}
		c.logger.Info(c.ctx, "server reconnection failed due to "+err.Error())
		time.Sleep(500)
	}
}
