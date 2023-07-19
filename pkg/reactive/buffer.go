package reactive

import (
	"context"
	"errors"
)

type Buffer[T any] struct {
	sinkEvents *SinkEvents[T]
	limit      int
	Output     []T
	Done       bool
}

func NewBuffer[T any](maxCount int) *Buffer[T] {
	return &Buffer[T]{
		sinkEvents: &SinkEvents[T]{},
		limit:      maxCount,
	}
}

func (s *Buffer[T]) Write(ctx context.Context, value T) error {
	if s.Done {
		return Done
	}
	if len(s.Output) >= s.limit {
		return Full
	}
	s.Output = append(s.Output, value)
	return nil
}

func (s *Buffer[T]) Finish(ctx context.Context) error {
	return errors.New("todo")
}

func (s *Buffer[T]) SinkEvents() *SinkEvents[T] {
	return s.sinkEvents
}

func (s *Buffer[T]) Resume(ctx context.Context) error {
	return errors.New("todo")
}
