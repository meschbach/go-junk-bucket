package reactive

import (
	"context"
	"errors"
)

type BufferState uint8

const (
	BufferInit BufferState = iota
	BufferFinished
)

type Buffer[T any] struct {
	state        BufferState
	sinkEvents   *SinkEvents[T]
	sourceEvents *SourceEvents[T]
	limit        int
	Output       []T
}

func NewBuffer[T any](maxCount int) *Buffer[T] {
	return &Buffer[T]{
		sinkEvents: &SinkEvents[T]{},
		limit:      maxCount,
	}
}

func (s *Buffer[T]) Write(ctx context.Context, value T) error {
	switch s.state {
	case BufferFinished:
		return Done
	}
	if len(s.Output) >= s.limit {
		return Full
	}
	s.Output = append(s.Output, value)
	return nil
}

func (s *Buffer[T]) Finish(ctx context.Context) error {
	switch s.state {
	case BufferFinished:
		return nil
	}

	s.state = BufferFinished
	return nil
}

func (s *Buffer[T]) SinkEvents() *SinkEvents[T] {
	return s.sinkEvents
}

func (s *Buffer[T]) Resume(ctx context.Context) error {
	return errors.New("todo")
}

func (s *Buffer[T]) SourceEvents() *SourceEvents[T] {
	return s.sourceEvents
}

func (s *Buffer[T]) ReadSlice(ctx context.Context, to []T) (int, error) {
	switch s.state {
	case BufferFinished:
		if s.Output == nil {
			return 0, End
		}
	}

	if s.Output == nil {
		return 0, nil
	}

	count := copy(to, s.Output)
	if count == len(s.Output) {
		s.Output = nil
	} else {
		s.Output = s.Output[count:]
	}
	return count, nil
}
