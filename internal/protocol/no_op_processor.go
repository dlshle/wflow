package protocol

import (
	"context"

	"github.com/dlshle/wflow/proto"
)

func NoOpProcessor(ctx context.Context, c GeneralConnection, m *proto.Message) error {
	return nil
}
