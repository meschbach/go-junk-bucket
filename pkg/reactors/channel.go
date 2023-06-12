package reactors

import "context"

// ChannelEvent is an opaque handle for relaying events back through the reactor from a channel receive.
type ChannelEvent struct {
	op TickEventFunc
}

// Channel is a Reactor implementation utilizing a chan as a work queue for execution.  When the driving event
// loop is ready it should receive from the work queue and dispatch using the Tick method.
type Channel struct {
	workQueue chan ChannelEvent
}

// NewChannel creates a new reactor with the specified queueSize.  It is recommended for queueSize to be greater
// than 1 to avoid synchronous hand-offs between goroutines or deadlocking.
func NewChannel(queueSize int) (*Channel, <-chan ChannelEvent) {
	workQueue := make(chan ChannelEvent, queueSize)
	return &Channel{workQueue: workQueue}, workQueue
}

func (c *Channel) Done() {
	close(c.workQueue)
}

func (c *Channel) ScheduleFunc(ctx context.Context, operation TickEventFunc) {
	select {
	case c.workQueue <- ChannelEvent{op: operation}:
	case <-ctx.Done():
	}
}

func (c *Channel) Tick(ctx context.Context, event ChannelEvent) error {
	return InvokeOp(ctx, c, event.op)
}
