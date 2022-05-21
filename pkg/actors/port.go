package actors

import (
	"context"
	"time"
)

type Port interface {
	Pid() Pid
	ReceiveChannel() <-chan any
	Receive() any
	ReceiveTimeout(wait time.Duration) (any, error)
	ReceiveWith(ctx context.Context) (any, error)
	Tell(ctx context.Context, who Pid, what any)
	Log() Logger
	Close(ctx context.Context)
}
