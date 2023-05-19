package message_processors

import (
	"context"

	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/logs"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateUploadLogsProcessor(logsStore logs.Store) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		WrappedLogs := &proto.WrappedLogs{}
		err := gproto.Unmarshal(m.Payload, WrappedLogs)
		if err != nil {
			return err
		}
		err = logsStore.BulkPut(WrappedLogs.Logs)
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, nil)
	}
}
