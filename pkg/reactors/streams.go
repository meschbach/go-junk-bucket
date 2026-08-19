package reactors

import (
	"context"
	"errors"

	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

// StreamBetweenOpt configures a [StreamBetween].
type StreamBetweenOpt func(s *streamBetweenConfig)

type streamBetweenConfig struct {
	name string
}

func (s *streamBetweenConfig) init(opts []StreamBetweenOpt) {
	s.name = "between-boundary"
	for _, o := range opts {
		o(s)
	}
}

// WithStreamBetweenName sets the tracing name prefix for the stream's internal spans.
func WithStreamBetweenName(name string) StreamBetweenOpt {
	return func(s *streamBetweenConfig) {
		s.name = name
	}
}

// StreamBetween creates a [streams.Source] and [streams.Sink] that bridge
// inputSide and outputSide reactor boundaries.
//
// Events written to the returned sink are delivered to outputSide's reactor,
// and feedback flows back to inputSide's reactor. Both sides execute within
// their respective reactor boundaries, preserving single-threaded semantics.
//
// StreamBetween should generally be invoked by a coordinating builder common
// to both sides, since the returned stream must be passed between boundaries.
func StreamBetween[E any, I any, O any](ctx context.Context, inputSide Boundary[I], outputSide Boundary[O], opts ...StreamBetweenOpt) (streams.Source[E], streams.Sink[E], error) {
	//figure out options
	cfg := streamBetweenConfig{}
	cfg.init(opts)

	//
	port := streams.NewChannelPort[E](32)

	outputSource := port.Output
	inputSink := port.Input

	inputSink.Push = func(ctx context.Context) error {
		outputSide.ScheduleFunc(ctx, func(parent context.Context) error {
			ctx, span := tracing.Start(parent, cfg.name+".consumer.feedback")
			defer span.End()
			_, err := outputSource.PumpTick(ctx)
			// End is a terminal signal delivered via SourceEvents.End.
			// Don't propagate it as an error to the reactor boundary.
			if errors.Is(err, streams.End) {
				return nil
			}
			return err
		})
		return nil
	}

	//todo: feedback mechanism should be pluggable so we can avoid an extra goroutine
	go func() {
		for event := range port.Feedback {
			eventCopy := event
			inputSide.ScheduleFunc(ctx, func(parent context.Context) error {
				ctx, span := tracing.Start(parent, cfg.name+".producer.feedback")
				defer span.End()
				return inputSink.ConsumeEvent(ctx, eventCopy)
			})
		}
	}()

	return outputSource, inputSink, nil
}
