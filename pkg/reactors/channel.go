package reactors

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

// ChannelEvent is an opaque handle returned from a [Channel]'s work queue.
// Pass it to [Channel.Tick] to execute the scheduled operation.
type ChannelEvent[S any] struct {
	op      TickEventStateFunc[S]
	invoker trace.SpanContext
}

// Channel is a [Boundary] implementation backed by a buffered channel work queue.
// The owning goroutine reads events from the queue and dispatches them via [Channel.Tick].
type Channel[S any] struct {
	workQueue chan ChannelEvent[S]
}

// NewChannel creates a [Channel] with the specified work queue buffer size.
// A buffer size greater than 1 is recommended to avoid synchronous handoffs
// between goroutines or deadlocking.
func NewChannel[S any](queueSize int) (*Channel[S], <-chan ChannelEvent[S]) {
	workQueue := make(chan ChannelEvent[S], queueSize)
	return &Channel[S]{workQueue: workQueue}, workQueue
}

// Done closes the work queue, signaling that no more events will be scheduled.
func (c *Channel[S]) Done() {
	close(c.workQueue)
}

// ScheduleFunc queues operation for execution.
// If ctx is canceled before the operation is queued, it is dropped.
func (c *Channel[S]) ScheduleFunc(ctx context.Context, operation TickEventFunc) {
	c.ScheduleStateFunc(ctx, func(ctx context.Context, state S) error {
		return operation(ctx)
	})
}

// ScheduleStateFunc queues operation for execution.
// If ctx is canceled before the operation is queued, it is dropped.
func (c *Channel[S]) ScheduleStateFunc(ctx context.Context, operation TickEventStateFunc[S]) {
	invoker := trace.SpanContextFromContext(ctx)
	select {
	case c.workQueue <- ChannelEvent[S]{op: operation, invoker: invoker}:
	case <-ctx.Done():
	}
}

// ConsumeAll processes all pending events in the work queue until empty.
// Returns the number of events processed and any error from [Channel.Tick]
// or context cancellation.
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

// Tick executes event within the reactor with the given state.
// The event's original caller's span context is linked for tracing.
func (c *Channel[S]) Tick(ctx context.Context, event ChannelEvent[S], state S) error {
	tickContext := trace.ContextWithRemoteSpanContext(ctx, event.invoker)
	return InvokeStateOp[S](tickContext, c, state, event.op)
}
