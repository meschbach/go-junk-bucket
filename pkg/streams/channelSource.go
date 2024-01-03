package streams

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	//pendingData indicates a data request has already been sent but no data has been received
	pendingData bool
}

func NewChannelSource[T any](pipe <-chan T) *ChannelSource[T] {
	return &ChannelSource[T]{
		events:      &SourceEvents[T]{},
		pipe:        pipe,
		mode:        channelSourcePaused,
		pendingData: false,
	}
}

func (c *ChannelSource[T]) SourceEvents() *SourceEvents[T] {
	return c.events
}

func (c *ChannelSource[T]) ReadSlice(parent context.Context, to []T) (i int, err error) {
	targetSize := len(to)
	ctx, span := tracing.Start(parent, "ChannelSource.ReadSlice", trace.WithAttributes(attribute.Int("capacity", targetSize)))
	defer span.End()

	i = 0
	keepReading := true
	for keepReading {
		select {
		case <-ctx.Done():
			return i, ctx.Err()
		case value, open := <-c.pipe:
			if open {
				c.resetFeedback()
				span.AddEvent("channel value")
				to[i] = value
				i++
				keepReading = targetSize > i
			} else {
				span.AddEvent("channel closed")
				if c.feedback != nil {
					close(c.feedback)
				}
				return i, End
			}
		default:
			span.AddEvent("channel empty")
			span.SetAttributes(attribute.Int("read", i))
			if err := c.notifyFeedback(ctx); err != nil {
				return i, ctx.Err()
			}
			return i, UnderRun
		}
	}
	span.SetAttributes(attribute.Int("read", i))
	return i, nil
}

func (c *ChannelSource[T]) Resume(parent context.Context) error {
	ctx, span := tracing.Start(parent, "ChannelSource.Resume")
	defer span.End()

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
			c.resetFeedback()
			if err := c.events.Data.Emit(ctx, value); err != nil {
				if errors.Is(err, Full) {
					c.mode = channelSourcePaused
				}
			}
		} else {
			endError := c.events.End.Emit(ctx, c)
			if c.feedback != nil {
				close(c.feedback)
			}
			return endError
		}
	}
	return nil
}

func (c *ChannelSource[T]) PumpTick(parent context.Context) (count int, err error) {
	ctx, span := tracing.Start(parent, "ChannelSource.PumpTick", trace.WithAttributes(attribute.Bool("pendingData", c.pendingData)))
	defer span.End()

	count = 0
	hasData := true
	for c.mode == channelSourceFlowing && hasData {
		select {
		case <-ctx.Done():
			span.AddEvent("context done")
			return count, ctx.Err()
		case value, open := <-c.pipe:
			if open {
				span.AddEvent("channel-value")
				c.resetFeedback()
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
				span.AddEvent("channel-end")
				problem := c.events.End.Emit(ctx, c)
				if c.feedback != nil {
					close(c.feedback)
				}
				outErr := End
				if problem != nil {
					outErr = errors.Join(outErr, problem)
				}
				return count, outErr
			}
		default:
			span.AddEvent("channel empty")
			hasData = false
		}
	}
	span.AddEvent("done")
	if err := c.notifyFeedback(ctx); err != nil {
		return count, err
	}
	return count, nil
}

func (c *ChannelSource[T]) Pause(ctx context.Context) error {
	c.mode = channelSourcePaused
	return nil
}

func (c *ChannelSource[T]) resetFeedback() {
	c.pendingData = false
}

// notifyFeedback notifies the listening feedback channel we are ready for more data but only once per successful read.
// This presents a noisy consumer from consuming a ton of the producers time.
func (c *ChannelSource[T]) notifyFeedback(ctx context.Context) error {
	if c.feedback == nil || c.pendingData {
		return nil
	}
	c.pendingData = true
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.feedback <- 1:
	default: //full?
	}
	return nil
}
