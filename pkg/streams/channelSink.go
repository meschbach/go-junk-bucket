package streams

import (
	"context"
	"errors"
)

// ChannelSink bridges a Golang `chan` into the bridge interface.  Due to the nature of channels not having feedback
// mechanisms some events will not fire as expected without additional aid.
type ChannelSink[T any] struct {
	events *SinkEvents[T]
	target chan<- T
}

func NewChannelSink[T any](target chan<- T) *ChannelSink[T] {
	return &ChannelSink[T]{
		events: &SinkEvents[T]{},
		target: target,
	}
}

func (c *ChannelSink[T]) Write(ctx context.Context, v T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.target <- v:
		elements := len(c.target)
		total := cap(c.target)
		if elements == total {
			return Full
		}
		return nil
	default:
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

func (c *ChannelSink[T]) ConsumeEvent(ctx context.Context, e channelPortFeedback) error {
	if err := c.events.Drained.Emit(ctx, c); err != nil {
		if errors.Is(err, Full) {
			return nil
		}
		return err
	}
	return nil
}
