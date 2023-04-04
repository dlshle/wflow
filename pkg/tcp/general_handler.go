package tcp

import (
	"context"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/notification"
	"github.com/dlshle/gts"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func createHandler(messageHandler MessageHandler, eventEmitter notification.WRNotificationEmitter[*proto.Message]) func(ctx context.Context, c gts.Connection, b []byte) {
	logger := logging.GlobalLogger.WithPrefix("[MessageHandler]")
	return func(ctx context.Context, c gts.Connection, b []byte) {
		m := &proto.Message{}
		err := gproto.Unmarshal(b, m)
		if err != nil {
			logger.Error(ctx, "failed to unmarshall message")
			return
		}
		ctx = logging.WrapCtx(ctx, "message_id", m.Id)
		logger.Info(ctx, "received message "+m.String())
		eventEmitter.Notify(m.Id, m)
		err = messageHandler.Handle(ctx, NewGeneralConnection(c), m)
		if err != nil {
			logger.Error(ctx, "failed to process message due to "+err.Error())
			return
		}
	}
}
