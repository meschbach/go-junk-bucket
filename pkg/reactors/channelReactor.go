package reactors

import "context"

// ChannelReactorEvent is an opaque handle for relaying events back through the reactor from a channel receive.
type ChannelReactorEvent struct {
	op TickEventFunc
}

// ChannelReactor is a Reactor implementation utilizing a chan as a work queue for execution.  When the driving event
// loop is ready it should receive from the work queue and dispatch using the Tick method.
type ChannelReactor struct {
	workQueue chan ChannelReactorEvent
}

// NewChannelReactor creates a new reactor with the specified queueSize.  It is recommended for queueSize to be greater
// than 1 to avoid synchronous hand-offs between goroutines or deadlocking.
func NewChannelReactor(queueSize int) (*ChannelReactor, <-chan ChannelReactorEvent) {
	workQueue := make(chan ChannelReactorEvent, queueSize)
	return &ChannelReactor{workQueue: workQueue}, workQueue
}

func (c *ChannelReactor) Done() {
	close(c.workQueue)
}

func (c *ChannelReactor) ScheduleFunc(ctx context.Context, operation TickEventFunc) {
	select {
	case c.workQueue <- ChannelReactorEvent{op: operation}:
	case <-ctx.Done():
	}
}

func (c *ChannelReactor) Tick(ctx context.Context, event ChannelReactorEvent) error {
	return InvokeOp(ctx, c, event.op)
}
