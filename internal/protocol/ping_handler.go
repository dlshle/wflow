package protocol

import (
	"context"

	"github.com/dlshle/wflow/proto"
)

func HandlePing(ctx context.Context, c GeneralConnection, m *proto.Message) error {
	m.Type = proto.Type_PONG
	return c.Send(m)
}
