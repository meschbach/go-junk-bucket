package streams

import (
	"context"
)

// SliceAccumulator is a Sink for storing T in Output.  Once Finish is called OnFinished handlers will be notified.
type SliceAccumulator[T any] struct {
	Output []T
	Done   bool
	Events *SinkEvents[T]
}

func NewSliceAccumulator[T any]() *SliceAccumulator[T] {
	return &SliceAccumulator[T]{
		Events: &SinkEvents[T]{},
	}
}

func (s *SliceAccumulator[T]) Write(ctx context.Context, value T) error {
	if s.Done {
		return Done
	}
	s.Output = append(s.Output, value)
	return nil
}

func (s *SliceAccumulator[T]) Finish(ctx context.Context) error {
	if s.Done {
		return nil
	}
	s.Done = true
	return s.Events.OnFinished.Emit(ctx, s)
}

func (s *SliceAccumulator[T]) SinkEvents() *SinkEvents[T] {
	return s.Events
}

func (s *SliceAccumulator[T]) Resume(ctx context.Context) error {
	return s.Events.OnDrain.Emit(ctx, s)
}
