package contrib

import (
	"context"
	"os/exec"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/proto"
)

func NewShellActivity() activity.WorkerActivity {
	description := "An activity to execute shell commands on the worker node"
	return activity.NewWorkerActivity(&proto.Activity{
		Name:        "shell activity",
		Description: &description,
	}, handler)
}

func handler(ctx context.Context, logger logging.Logger, input []byte) (output []byte, err error) {
	cmd := exec.Command(string(input))
	return cmd.Output()
}
