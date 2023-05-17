package protocol

import (
	"context"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/notification"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func createHandler(messageHandler MessageHandler, eventEmitter notification.WRNotificationEmitter[*proto.Message]) func(ctx context.Context, c GeneralConnection, b []byte) {
	logger := logging.GlobalLogger.WithPrefix("[MessageHandler]")
	return func(ctx context.Context, c GeneralConnection, b []byte) {
		m := &proto.Message{}
		err := gproto.Unmarshal(b, m)
		if err != nil {
			logger.Error(ctx, "failed to unmarshall message")
			handleError(c, m, err)
			return
		}
		ctx = logging.WrapCtx(ctx, "message_id", m.Id)
		logger.Info(ctx, "received message "+m.String())
		processed := eventEmitter.Notify(m.Id, m)
		if m.Type == proto.Type_RESPONSE || processed {
			return
		}
		err = messageHandler.Handle(ctx, c, m)
		if err != nil {
			logger.Error(ctx, "failed to process message due to "+err.Error())
			handleError(c, m, err)
			return
		}
	}
}

func handleError(c GeneralConnection, m *proto.Message, err error) {
	m.Type = proto.Type_RESPONSE
	m.Payload = []byte(err.Error())
	m.Status = proto.Status_INTERNAL
	c.Send(m)
}
