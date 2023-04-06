package activity

import "context"

type ActivityHandler func(context.Context, []byte) ([]byte, error)
