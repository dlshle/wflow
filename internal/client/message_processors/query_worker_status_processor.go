package message_processors

import (
	"context"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/wflow/internal/client/job"
	"github.com/dlshle/wflow/pkg/protocol"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateQueryWorkerStatusProcessor(workerID string, jobManager job.JobManager) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		worker := &proto.Worker{}
		err := gproto.Unmarshal(m.Payload, worker)
		if err != nil {
			return errors.Error("failed to parse query worker request due to " + err.Error())
		}
		if worker.Id != workerID {
			return errors.Error("incorrect worker ID " + worker.Id + " for query worker status")
		}
		workerData, err := gproto.Marshal(jobManager.WorkerInfo())
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, workerData)
	}
}
