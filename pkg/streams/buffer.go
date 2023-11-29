package streams

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type bufferState uint8

func (b bufferState) String() string {
	switch b {
	case bufferPaused:
		return "paused"
	case bufferFlowing:
		return "flowing"
	case bufferFinished:
		return "finished"
	default:
		panic("unknown buffer state")
	}
}

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

func (s *Buffer[T]) Write(parent context.Context, value T) error {
	ctx, span := tracing.Start(parent, "Buffer.Write", trace.WithAttributes(attribute.Stringer("state", s.state), attribute.Stringer("Buffer", ptrFormattedStringer[Buffer[T]]{s})))

	switch s.state {
	case bufferFinished:
		return Done
	case bufferPaused:
		if len(s.Output) >= s.limit {
			span.AddEvent("full")
			return Full
		}
		s.Output = append(s.Output, value)
		span.AddEvent("buffered")
	case bufferFlowing:
		span.AddEvent("flowing")
		if err := s.sourceEvents.Data.Emit(ctx, value); err != nil {
			if errors.Is(err, Full) {
				span.AddEvent("destinations-full")
				s.Output = append(s.Output, value)
				s.state = bufferPaused
			} else {
				span.SetStatus(codes.Error, "source dispatch failed")
				span.RecordError(err)
				return err
			}
		}
	}
	return nil
}

// Finish will transition the buffer into FinalFlush mode until all elements are read, at which point the buffer will
// the transition into Finished.
func (s *Buffer[T]) Finish(ctx context.Context) error {
	switch s.state {
	case bufferFinished:
		return nil
	}

	//todo: intermediate state where we are draining the buffers
	s.state = bufferFinished
	//todo: test this is dispatched
	sourceEmitErrors := s.sourceEvents.End.Emit(ctx, s)
	sinkEmitErrors := s.sinkEvents.OnFinished.Emit(ctx, s)
	return errors.Join(sourceEmitErrors, sinkEmitErrors)
}

func (s *Buffer[T]) SinkEvents() *SinkEvents[T] {
	return s.sinkEvents
}

func (s *Buffer[T]) Pause(ctx context.Context) error {
	switch s.state {
	case bufferFinished:
		return End
	case bufferFlowing:
		s.state = bufferPaused
	}
	return nil
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

type ptrFormattedStringer[T any] struct {
	obj *T
}

func (p ptrFormattedStringer[T]) String() string {
	return fmt.Sprintf("%p", p.obj)
}
