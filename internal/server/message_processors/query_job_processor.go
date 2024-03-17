package message_processors

import (
	"context"

	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/job"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateQueryJobProcessor(manager worker.Manager, jobHandler job.Handler) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		job := &proto.Job{}
		err := gproto.Unmarshal(m.Payload, job)
		if err != nil {
			return err
		}
		serverJob, err := jobHandler.Get(job.Id)
		manager.GetJob(ctx, job.Id)
		if err != nil {
			return err
		}
		marshalledServerJob, err := gproto.Marshal(serverJob)
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, marshalledServerJob)
	}
}
