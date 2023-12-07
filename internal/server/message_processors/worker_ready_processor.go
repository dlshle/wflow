package message_processors

import (
	"context"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateWorkerReadyProcessor(manager worker.Manager) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		logging.GlobalLogger.Debugf(ctx, "received worker ready message %v from gc %v", m, gc.ConnGID())
		worker := &proto.Worker{}
		err := gproto.Unmarshal(m.Payload, worker)
		if err != nil {
			return err
		}
		worker.WorkerIp = gc.Address()
		err = manager.HandleWorkerUpdate(ctx, worker)
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, nil)
	}
}
