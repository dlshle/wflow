package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/proto"
)

type loggingMessageHandler struct {
	id string
}

func newLoggingMessageHandler(id string) MessageHandler {
	return &loggingMessageHandler{id}
}

func (h *loggingMessageHandler) Handle(ctx context.Context, gc GeneralConnection, m *proto.Message) error {
	logging.GlobalLogger.Debugf(ctx, "logging message handler for {%s} received message %v", h.id, m)
	if m.Id == "NeedsResponse" {
		logging.GlobalLogger.Debugf(ctx, "{%s}: received message %v and will reply", h.id, m)
		return gc.Respond(m, proto.Status_OK, []byte("response"))
	}
	logging.GlobalLogger.Debugf(ctx, "{%s}: received message %v", h.id, m)
	return nil
}

func TestProtocol(t *testing.T) {
	s := NewTCPServer("server", "localhost", 14514, newLoggingMessageHandler("server"))
	go func() {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	}()
	c := NewTCPClient("client", "localhost", 14514, newLoggingMessageHandler("client"), []*proto.Activity{}, func(sc ServerConnection, s *proto.Server) {
		logging.GlobalLogger.Debugf(context.Background(), "protocol exchanged with server %v", s)
	})
	serverConn, err := c.Connect()
	if err != nil {
		panic(err)
	}
	err = serverConn.Send(&proto.Message{Id: "1", Type: proto.Type_PING, Payload: []byte("1st payload")})
	if err != nil {
		panic(err)
	}
	resp, err := serverConn.Request(&proto.Message{Id: "NeedsResponse", Type: proto.Type_CUSTOM})
	logging.GlobalLogger.Debugf(context.Background(), "sent NeedsResponse msg")
	if err != nil {
		panic(err)
	}
	logging.GlobalLogger.Debugf(context.Background(), "client received response: %v", resp)
	time.Sleep(15 * time.Second)
	// logging.GlobalLogger.Debugf(context.Background(), "received response %v", resp)
	err = serverConn.Close()
	if err != nil {
		panic(err)
	}
	panic("err")
}
