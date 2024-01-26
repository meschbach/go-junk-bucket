package reactors

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

// ChannelEvent is an opaque handle for relaying events back through the reactor from a channel receive.
type ChannelEvent[S any] struct {
	op      TickEventStateFunc[S]
	invoker trace.SpanContext
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
	invoker := trace.SpanContextFromContext(ctx)
	select {
	case c.workQueue <- ChannelEvent[S]{op: operation, invoker: invoker}:
	case <-ctx.Done():
	}
}

// ConsumeAll will consume all pending messages from the work queue until empty then return.  If the invoking context
// is canceled prior to completion then routine will exit early.
func (c *Channel[S]) ConsumeAll(ctx context.Context, state S) (int, error) {
	count := 0
	for {
		select {
		case e := <-c.workQueue:
			count++
			if err := c.Tick(ctx, e, state); err != nil {
				return count, err
			}
		case <-ctx.Done():
			return count, ctx.Err()
		default:
			return count, nil
		}
	}
}

// Tick will dispatch the requested event within the context of this reactor.
func (c *Channel[S]) Tick(ctx context.Context, event ChannelEvent[S], state S) error {
	tickContext := trace.ContextWithRemoteSpanContext(ctx, event.invoker)
	return InvokeStateOp[S](tickContext, c, state, event.op)
}
