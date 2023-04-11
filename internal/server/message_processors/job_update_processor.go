package messageprocessors

import (
	"context"

	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/job"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateJobUpdateProcessor(manager worker.Manager, jobHandler job.Handler) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		job := &proto.JobReport{}
		err := gproto.Unmarshal(m.Payload, job)
		if err != nil {
			return err
		}
		err = jobHandler.Put(job)
		if err != nil {
			return err
		}
		worker, err := manager.QueryRemoteWorker(ctx, job.WorkerId)
		if err != nil {
			return err
		}
		err = manager.HandleWorkerUpdate(ctx, worker)
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, nil)
	}
}
