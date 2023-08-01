package streams

import (
	"context"
	"errors"
)

type bufferState uint8

const (
	bufferPaused bufferState = iota
	bufferFlowing
	bufferFinished
)

// Buffer will hold up to a limit of elements before placing back pressure on a writer.  Usable as a source also.
type Buffer[T any] struct {
	state        bufferState
	sinkEvents   *SinkEvents[T]
	sourceEvents *SourceEvents[T]
	limit        int
	Output       []T
}

// NewBuffer creates a new Buffer with the specified maxCount as the limit
func NewBuffer[T any](maxCount int) *Buffer[T] {
	return &Buffer[T]{
		state:        bufferPaused,
		sinkEvents:   &SinkEvents[T]{},
		sourceEvents: &SourceEvents[T]{},
		limit:        maxCount,
	}
}

func (s *Buffer[T]) Write(ctx context.Context, value T) error {
	switch s.state {
	case bufferFinished:
		return Done
	case bufferPaused:
		if len(s.Output) >= s.limit {
			return Full
		}
		s.Output = append(s.Output, value)
	case bufferFlowing:
		if err := s.sourceEvents.Data.Emit(ctx, value); err != nil {
			if errors.Is(err, Full) {
				s.Output = append(s.Output, value)
				s.state = bufferPaused
			} else {
				return err
			}
		}
	}
	return nil
}

func (s *Buffer[T]) Finish(ctx context.Context) error {
	switch s.state {
	case bufferFinished:
		return nil
	}

	s.state = bufferFinished
	return nil
}

func (s *Buffer[T]) SinkEvents() *SinkEvents[T] {
	return s.sinkEvents
}

func (s *Buffer[T]) Resume(ctx context.Context) error {
	switch s.state {
	case bufferFinished:
		return End
	case bufferFlowing:
		return nil
	}
	for {
		if len(s.Output) == 0 {
			if err := s.sinkEvents.OnDrain.Emit(ctx, s); err != nil {
				return err
			}
			if len(s.Output) == 0 {
				s.state = bufferFlowing
				return nil
			}
		}
		e := s.Output[0]
		if err := s.sourceEvents.Data.Emit(ctx, e); err == nil {
			s.Output = s.Output[1:]
		} else {
			return err
		}
	}
}

func (s *Buffer[T]) SourceEvents() *SourceEvents[T] {
	return s.sourceEvents
}

func (s *Buffer[T]) ReadSlice(ctx context.Context, to []T) (int, error) {
	switch s.state {
	case bufferFinished:
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
