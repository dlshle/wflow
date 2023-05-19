package protocol

import (
	"context"
	"time"

	"github.com/dlshle/gommon/errors"
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
	return c.c.Close()
}

func (c *generalConnection) requestWithTimeout(timeoutInMs int, m *proto.Message) (resp *proto.Message, err error) {
	channel := make(chan *proto.Message, 1)
	disposable, err := c.notificationCenter.On(m.Id, func(m *proto.Message) {
		channel <- m
	})
	if err != nil {
		return nil, err
	}
	err = c.Send(m)
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(time.Millisecond * time.Duration(timeoutInMs)):
		err = errors.Error("request " + m.Id + " timedout")
		disposable()
		return
	case resp = <-channel:
		disposable()
		return
	}
}
