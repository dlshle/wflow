package message_processors

import (
	"context"

	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/job"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateScheduleJobProcessor(manager worker.Manager, jobHandler job.Handler) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		job := &proto.Job{}
		err := gproto.Unmarshal(m.Payload, job)
		if err != nil {
			return err
		}
		dispatchedJob, err := manager.ScheduleJob(ctx, job)
		if err != nil {
			return err
		}
		marshalledDispatchedJob, err := gproto.Marshal(dispatchedJob)
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, marshalledDispatchedJob)
	}
}
