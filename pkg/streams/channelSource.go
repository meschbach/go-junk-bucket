package streams

import (
	"context"
	"errors"
)

type channelSourceMode uint8

const (
	channelSourcePaused channelSourceMode = iota
	channelSourceFlowing
)

type ChannelSource[T any] struct {
	events   *SourceEvents[T]
	pipe     <-chan T
	feedback chan channelPortFeedback
	mode     channelSourceMode
}

func NewChannelSource[T any](pipe <-chan T) *ChannelSource[T] {
	return &ChannelSource[T]{
		events: &SourceEvents[T]{},
		pipe:   pipe,
		mode:   channelSourcePaused,
	}
}

func (c *ChannelSource[T]) SourceEvents() *SourceEvents[T] {
	return c.events
}

func (c *ChannelSource[T]) ReadSlice(ctx context.Context, to []T) (i int, err error) {
	i = 0
	keepReading := true
	targetSize := len(to)
	for keepReading {
		select {
		case <-ctx.Done():
			return i, ctx.Err()
		case value, open := <-c.pipe:
			if open {
				to[i] = value
				i++
				keepReading = targetSize > i
			} else {
				return i, End
			}
		default:
			if c.feedback != nil {
				select {
				case <-ctx.Done():
					return i, ctx.Err()
				case c.feedback <- channelPortFeedback(i):
					break
				default: //busy feedback?
				}
			}
			return i, UnderRun
		}
	}
	return i, nil
}

func (c *ChannelSource[T]) Resume(ctx context.Context) error {
	c.mode = channelSourceFlowing
	_, err := c.PumpTick(ctx)
	return err
}

func (c *ChannelSource[T]) WaitOnEvent(ctx context.Context) error {
	if c.mode != channelSourceFlowing {
		panic("waiting when not flowing")
		//return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case value, open := <-c.pipe:
		if open {
			if err := c.events.Data.Emit(ctx, value); err != nil {
				if errors.Is(err, Full) {
					c.mode = channelSourcePaused
				}
			}
		} else {
			//todo: end
			return nil
		}
	}
	return nil
}

func (c *ChannelSource[T]) PumpTick(ctx context.Context) (count int, err error) {
	count = 0
	hasData := true
	for c.mode == channelSourceFlowing && hasData {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		case value, open := <-c.pipe:
			if open {
				count++
				if err := c.events.Data.Emit(ctx, value); err != nil {
					// A consumer of our data is full and needs us to buffer.  We'll toggle back to paused mode.
					if errors.Is(err, Full) {
						c.mode = channelSourcePaused
						return count, nil
					}
					return count, err
				}
			} else {
				return count, End
			}
		default:
			hasData = false
		}
	}
	if c.feedback != nil {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		case c.feedback <- 1:
		default: //full?
		}
	}
	return count, nil
}

func (c *ChannelSource[T]) Pause(ctx context.Context) error {
	c.mode = channelSourcePaused
	return nil
}
