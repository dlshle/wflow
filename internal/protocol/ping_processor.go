package protocol

import (
	"context"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/proto"
)

func PingProcessor(ctx context.Context, c GeneralConnection, m *proto.Message) error {
	logging.GlobalLogger.Infof(ctx, "received ping request from %s", c.Address())
	m.Type = proto.Type_PONG
	return c.Send(m)
}
