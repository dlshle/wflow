package tcp

import (
	"context"

	"github.com/dlshle/gommon/async"
)

type tcpClient struct {
	ctx context.Context
	id  string

	asyncPool async.AsyncPool
}

// TODO: implement initial protocol exchange for c/s com
