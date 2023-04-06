package job

import (
	"context"

	"github.com/dlshle/wflow/proto"
)

type cancellableJobReport struct {
	jobCtx     context.Context
	cancelFunc func()
	*proto.JobReport
}
