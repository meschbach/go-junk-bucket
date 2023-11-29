package streams

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type bufferReadState uint8

func (b bufferReadState) String() string {
	switch b {
	case bufferPaused:
		return "paused"
	case bufferFlowing:
		return "flowing"
	case bufferDone:
		return "done"
	default:
		panic(fmt.Sprintf("unknown buffer read state %d", b))
	}
}

const (
	bufferPaused bufferReadState = iota
	bufferFlowing
	bufferDone
)

type bufferWriteState uint8

const (
	bufferWritable bufferWriteState = iota
	bufferFinishing
	bufferFinished
)

func (b bufferWriteState) String() string {
	switch b {
	case bufferWritable:
		return "writable"
	case bufferFinishing:
		return "finishing"
	case bufferFinished:
		return "finished"
	default:
		panic(fmt.Sprintf("unknown buffer write state %d", b))
	}
}

// Buffer will hold up to a limit of elements before placing back pressure on a writer.  Usable as a source also.
type Buffer[T any] struct {
	readState    bufferReadState
	writeState   bufferWriteState
	sinkEvents   *SinkEvents[T]
	sourceEvents *SourceEvents[T]
	limit        int
	Output       []T
}

// NewBuffer creates a new Buffer with the specified maxCount as the limit
func NewBuffer[T any](maxCount int) *Buffer[T] {
	return &Buffer[T]{
		readState:    bufferPaused,
		writeState:   bufferWritable,
		sinkEvents:   &SinkEvents[T]{},
		sourceEvents: &SourceEvents[T]{},
		limit:        maxCount,
	}
}

func (s *Buffer[T]) Write(parent context.Context, value T) error {
	ctx, span := tracing.Start(parent, "Buffer.Write", trace.WithAttributes(attribute.Stringer("state.write", s.writeState), attribute.Stringer("Buffer", ptrFormattedStringer[Buffer[T]]{s})))

	switch s.writeState {
	case bufferWritable:
		switch s.readState {
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
					s.readState = bufferPaused
				} else {
					span.SetStatus(codes.Error, "source dispatch failed")
					span.RecordError(err)
					return err
				}
			}
		}
	case bufferFinishing:
		return Done
	case bufferFinished:
		return Done
	}
	return nil
}

// Finish will transition the buffer into FinalFlush mode until all elements are read, at which point the buffer will
// the transition into Finished.
func (s *Buffer[T]) Finish(ctx context.Context) error {
	switch s.writeState {
	case bufferFinished:
		return nil
	case bufferFinishing:
		return nil
	case bufferWritable:
		return s.startFinalDrain(ctx)
	default:
		panic(fmt.Sprintf("unknown write state: %s", s.writeState))
	}
}

func (s *Buffer[T]) SinkEvents() *SinkEvents[T] {
	return s.sinkEvents
}

func (s *Buffer[T]) Pause(ctx context.Context) error {
	switch s.readState {
	case bufferFlowing:
		s.readState = bufferPaused
	}
	return nil
}

func (s *Buffer[T]) Resume(ctx context.Context) error {
	switch s.readState {
	case bufferDone:
		return End
	case bufferFlowing:
		return nil
	}
	for {
		if len(s.Output) == 0 {
			if err := s.sinkEvents.OnDrain.Emit(ctx, s); err != nil {
				return err
			}
			//check again, as Drain emission may have inserted more elements into our buffer
			if len(s.Output) == 0 {
				switch s.writeState {
				case bufferFinishing:
					return s.finishedFinalDrain(ctx)
				default:
					s.readState = bufferFlowing
				}
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
	switch s.readState {
	case bufferDone:
		return 0, End
	}

	if s.Output == nil {
		return 0, nil
	}

	count := copy(to, s.Output)
	if count == len(s.Output) {
		s.Output = nil
		switch s.writeState {
		case bufferFinishing:
			return count, s.finishedFinalDrain(ctx)
		}
	} else {
		s.Output = s.Output[count:]
	}

	return count, nil
}

func (s *Buffer[T]) startFinalDrain(ctx context.Context) error {
	s.writeState = bufferFinishing
	finishingError := s.sinkEvents.OnFinishing.Emit(ctx, s)

	var finalDrainError error
	if len(s.Output) == 0 {
		finalDrainError = s.finishedFinalDrain(ctx)
	}
	return errors.Join(finishingError, finalDrainError)
}

func (s *Buffer[T]) finishedFinalDrain(ctx context.Context) error {
	if len(s.Output) != 0 {
		panic("finishing final drain before empty buffer")
	}

	s.writeState = bufferFinished
	finishedError := s.sinkEvents.OnFinished.Emit(ctx, s)

	s.readState = bufferDone
	endError := s.sourceEvents.End.Emit(ctx, s)

	return errors.Join(finishedError, endError)
}

// todo: move
type ptrFormattedStringer[T any] struct {
	obj *T
}

func (p ptrFormattedStringer[T]) String() string {
	return fmt.Sprintf("%p", p.obj)
}
