package message_processors

import (
	"context"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/wflow/internal/client/job"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

func CreateQueryJobProcessor(jobManager job.JobManager) protocol.MessageProcessor {
	return func(ctx context.Context, gc protocol.GeneralConnection, m *proto.Message) error {
		job := &proto.Job{}
		err := gproto.Unmarshal(m.Payload, job)
		if err != nil {
			return errors.Error("failed to parse query job request due to " + err.Error())
		}
		report, err := jobManager.Job(job.Id)
		if err != nil {
			return err
		}
		reportData, err := gproto.Marshal(report)
		if err != nil {
			return err
		}
		return gc.Respond(m, proto.Status_OK, reportData)
	}
}
