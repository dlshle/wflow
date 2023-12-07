package protocol

import (
	"context"
	"time"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/notification"
	"github.com/dlshle/gts"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type MessageProcessor func(context.Context, GeneralConnection, *proto.Message) error

type GeneralConnection interface {
	ConnGID() string
	Send(*proto.Message) error
	Respond(m *proto.Message, status proto.Status, payload []byte) error
	Request(*proto.Message) (*proto.Message, error)
	Address() string
	IsActive() bool
	Close() error
}

type generalConnection struct {
	c                  gts.Connection
	timeoutInMs        int
	notificationCenter notification.WRNotificationEmitter[*proto.Message]
	gid                string
}

func NewGeneralConnection(c gts.Connection, notificationCenter notification.WRNotificationEmitter[*proto.Message], requestTimeoutMs int) GeneralConnection {
	return &generalConnection{c: c, timeoutInMs: requestTimeoutMs, notificationCenter: notificationCenter, gid: utils.RandomUUID()}
}

func (c *generalConnection) ConnGID() string {
	return c.gid
}

func (c *generalConnection) Address() string {
	return c.c.Address()
}

func (c *generalConnection) Send(m *proto.Message) error {
	data, err := gproto.Marshal(m)
	if err != nil {
		return err
	}
	return c.c.Write(data)
}

func (c *generalConnection) Respond(m *proto.Message, status proto.Status, payload []byte) error {
	m.Type = proto.Type_RESPONSE
	m.Status = status
	if payload != nil {
		m.Payload = payload
	} else {
		m.Payload = []byte{1}
	}
	return c.Send(m)
}

func (c *generalConnection) Request(m *proto.Message) (*proto.Message, error) {
	if !c.c.IsLive() {
		return nil, errors.Error("connection isn't established or is lost")
	}
	return c.requestWithTimeout(c.timeoutInMs, m)
}

func (c *generalConnection) Close() error {
	logging.GlobalLogger.Info(context.Background(), "closing connection "+c.c.String())
	return c.c.Close()
}

func (c *generalConnection) IsActive() bool {
	return c.Send(&proto.Message{Type: proto.Type_PING, Payload: []byte{1}}) == nil
}

func (c *generalConnection) requestWithTimeout(timeoutInMs int, m *proto.Message) (resp *proto.Message, err error) {
	startTime := time.Now()
	channel := make(chan *proto.Message, 1)
	disposable, err := c.notificationCenter.On(m.Id, func(m *proto.Message) {
		logging.GlobalLogger.Debugf(context.Background(), "received response for message %s", m.Id)
		channel <- m
	})
	if err != nil {
		logging.GlobalLogger.Errorf(context.Background(), "unable to register notification for message %s due to %s", m.Id, err.Error())
		return nil, err
	}
	defer disposable()
	err = c.Send(m)
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(time.Millisecond * time.Duration(timeoutInMs)):
		logging.GlobalLogger.Errorf(context.Background(), "request %v timedout for %v, but actual time-elapse is %v", m, time.Millisecond*time.Duration(timeoutInMs), time.Since(startTime))
		err = errors.Error("request " + m.Id + " timedout")
		return
	case resp = <-channel:
		return
	}
}
