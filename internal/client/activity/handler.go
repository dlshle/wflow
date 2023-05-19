package activity

import (
	"context"

	"github.com/dlshle/gommon/logging"
)

type ActivityHandler func(context.Context, logging.Logger, []byte) ([]byte, error)
