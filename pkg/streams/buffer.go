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
	traceConfig  bufferTraceOpts
}

type BufferOpt[T any] func(b *Buffer[T])

func WithBufferTracePrefix[T any](prefix string) BufferOpt[T] {
	return func(b *Buffer[T]) {
		b.traceConfig.writeSpan = prefix + ".Write"
		b.traceConfig.readSlice = prefix + ".ReadSlice"
	}
}

type bufferTraceOpts struct {
	writeSpan string
	readSlice string
}

// NewBuffer creates a new Buffer with the specified maxCount as the limit
func NewBuffer[T any](maxCount int, opts ...BufferOpt[T]) *Buffer[T] {
	b := &Buffer[T]{
		readState:    bufferPaused,
		writeState:   bufferWritable,
		sinkEvents:   &SinkEvents[T]{},
		sourceEvents: &SourceEvents[T]{},
		limit:        maxCount,
		traceConfig: bufferTraceOpts{
			writeSpan: "Buffer.Write",
			readSlice: "Buffer.ReadSlice",
		},
	}
	for _, o := range opts {
		o(b)
	}

	return b
}

func (s *Buffer[T]) writePausedBuffer(ctx context.Context, value T) error {
	//Already full, outright reject the overflow
	if len(s.Output) >= s.limit {
		return Overflow
	}

	s.Output = append(s.Output, value)
	if len(s.Output) >= s.limit {
		if err := s.sinkEvents.Full.Emit(ctx, s); err != nil {
			return err
		}
		return Full
	}
	return nil
}

func (s *Buffer[T]) Write(parent context.Context, value T) error {
	ctx, span := tracing.Start(parent, s.traceConfig.writeSpan, trace.WithAttributes(
		attribute.Stringer("state.write", s.writeState),
		attribute.Stringer("state.read", s.readState),
	))
	defer span.End()

	switch s.writeState {
	case bufferWritable:
		switch s.readState {
		case bufferPaused:
			return s.writePausedBuffer(ctx, value)
		case bufferFlowing:
			span.AddEvent("flowing")
			if err := s.sourceEvents.Data.Emit(ctx, value); err != nil {
				if errors.Is(err, Full) {
					s.readState = bufferPaused
					if len(s.Output) == s.limit {
						return Full
					} else {
						return nil
					}
				} else if errors.Is(err, Overflow) {
					panic("transitioned to overflow without full")
				} else {
					span.SetStatus(codes.Error, "source dispatch failed")
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
	case bufferPaused: //transition out
	}
	for {
		if len(s.Output) == 0 {
			if err := s.sinkEvents.Drained.Emit(ctx, s); err != nil {
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
		s.Output = s.Output[1:]
		if err := s.sourceEvents.Data.Emit(ctx, e); err != nil {
			return err
		}
	}
}

func (s *Buffer[T]) SourceEvents() *SourceEvents[T] {
	return s.sourceEvents
}

func (s *Buffer[T]) ReadSlice(parent context.Context, to []T) (int, error) {
	ctx, span := tracing.Start(parent, s.traceConfig.readSlice, trace.WithAttributes(
		attribute.Stringer("state.write", s.writeState),
		attribute.Stringer("state.read", s.readState),
	))
	defer span.End()

	//Are we in a valid state to continue?
	switch s.readState {
	case bufferDone:
		return 0, End
	case bufferFlowing: //programming error
		return 0, errors.New("flowing")
	case bufferPaused: //expected state to actually allow reads
	}

	countRead := 0
	for countRead < len(to) {
		//ensure we have not yet changed states.
		if s.readState != bufferPaused {
			break
		}

		//Buffer drained?  Solicit input!
		if len(s.Output) == 0 {
			span.AddEvent("buffer empty")
			if s.writeState == bufferWritable { //still writing?
				span.AddEvent("writable")
				if err := s.sinkEvents.Drained.Emit(ctx, s); err != nil {
					return countRead, err
				}
			}
			//do we still not have elements?
			if len(s.Output) == 0 {
				if countRead == 0 {
					span.AddEvent("under run")
					return 0, UnderRun
				} else {
					span.AddEvent("filled buffer")
					return countRead, nil
				}
			}
		}

		//We have data, read it into the target buffer
		copiedCount := copy(to[countRead:], s.Output)
		s.Output = s.Output[copiedCount:]
		countRead += copiedCount
	}
	if s.writeState == bufferWritable {
		span.AddEvent("writable")
		err := s.sinkEvents.Available.Emit(ctx, SinkAvailableEvent[T]{
			Space: s.limit - len(s.Output),
			Sink:  s,
		})
		return countRead, err
	} else {
		return countRead, nil
	}
}

// drainedOnRead emits the Drained event to notify possible listeners our buffer has under run.
func (s *Buffer[T]) drainedOnRead(ctx context.Context) error {
	if err := s.sinkEvents.Drained.Emit(ctx, s); err != nil {
		if errors.Is(err, Full) {
			err = nil
		}
		return err
	}
	if len(s.Output) > 0 {
		return nil
	} else {
		return UnderRun
	}
}

func (s *Buffer[T]) startFinalDrain(ctx context.Context) error {
	s.writeState = bufferFinishing
	finishingError := s.sinkEvents.Finishing.Emit(ctx, s)

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
	finishedError := s.sinkEvents.Finished.Emit(ctx, s)

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
