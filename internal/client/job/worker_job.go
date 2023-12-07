package job

import (
	"context"

	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/proto"
)

type workerJobReport struct {
	jobCtx     context.Context
	activity   activity.WorkerActivity
	cancelFunc func()
	*proto.JobReport
}
