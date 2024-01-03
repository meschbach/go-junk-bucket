package streams

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ChannelSink bridges a Golang `chan` into the bridge interface.  Due to the nature of channels not having feedback
// mechanisms some events will not fire as expected without additional aid.
type ChannelSink[T any] struct {
	events *SinkEvents[T]
	target chan<- T
	//Waiting indicates a consumer on the opposite side is waiting on a response
	waiting bool
	Push    func(ctx context.Context) error
}

func NewChannelSink[T any](target chan<- T) *ChannelSink[T] {
	return &ChannelSink[T]{
		events:  &SinkEvents[T]{},
		target:  target,
		waiting: false,
		Push: func(ctx context.Context) error {
			return nil
		},
	}
}

func (c *ChannelSink[T]) Write(parent context.Context, v T) error {
	ctx, span := tracing.Start(parent, "ChannelSink.Write", trace.WithAttributes(attribute.Bool("waiting", c.waiting)))
	defer span.End()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.target <- v:
		elements := len(c.target)
		total := cap(c.target)
		if c.waiting {
			if err := c.Push(ctx); err != nil {
				span.SetStatus(codes.Error, "failed push")
				return err
			}
		}
		if elements == total {
			span.AddEvent("full")
			return Full
		}
		return nil
	default:
		span.AddEvent("overflow")
		return Overflow
	}
}

func (c *ChannelSink[T]) Finish(ctx context.Context) error {
	finishingProblems := c.events.Finishing.Emit(ctx, c)
	close(c.target)
	finishedProblems := c.events.Finished.Emit(ctx, c)
	return errors.Join(finishingProblems, finishedProblems)
}

func (c *ChannelSink[T]) SinkEvents() *SinkEvents[T] {
	return c.events
}

func (c *ChannelSink[T]) Resume(ctx context.Context) error {
	//todo: the channel itself does not really support this -- figure out how
	return nil
}

func (c *ChannelSink[T]) ConsumeEvent(parent context.Context, e channelPortFeedback) error {
	ctx, span := tracing.Start(parent, "ChannelSink.ConsumeEvent")
	defer span.End()

	c.waiting = true
	if err := c.events.Drained.Emit(ctx, c); err != nil {
		if errors.Is(err, Full) {
			span.AddEvent("full")
			return nil
		}
		return err
	}
	return nil
}
