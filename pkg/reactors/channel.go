package reactors

import "context"

// ChannelEvent is an opaque handle for relaying events back through the reactor from a channel receive.
type ChannelEvent[S any] struct {
	op TickEventStateFunc[S]
}

// Channel is a Boundary implementation utilizing a chan as a work queue for execution.  When the driving event
// loop is ready it should receive from the work queue and dispatch using the Tick method.
type Channel[S any] struct {
	workQueue chan ChannelEvent[S]
}

// NewChannel creates a new reactor with the specified queueSize.  It is recommended for queueSize to be greater
// than 1 to avoid synchronous hand-offs between goroutines or deadlocking.
func NewChannel[S any](queueSize int) (*Channel[S], <-chan ChannelEvent[S]) {
	workQueue := make(chan ChannelEvent[S], queueSize)
	return &Channel[S]{workQueue: workQueue}, workQueue
}

func (c *Channel[S]) Done() {
	close(c.workQueue)
}

func (c *Channel[S]) ScheduleFunc(ctx context.Context, operation TickEventFunc) {
	c.ScheduleStateFunc(ctx, func(ctx context.Context, state S) error {
		return operation(ctx)
	})
}

func (c *Channel[S]) ScheduleStateFunc(ctx context.Context, operation TickEventStateFunc[S]) {
	select {
	case c.workQueue <- ChannelEvent[S]{op: operation}:
	case <-ctx.Done():
	}
}

func (c *Channel[S]) Tick(ctx context.Context, event ChannelEvent[S], state S) error {
	return InvokeStateOp[S](ctx, c, state, event.op)
}
