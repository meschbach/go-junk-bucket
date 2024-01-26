package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

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

func WithStreamBetweenName(name string) StreamBetweenOpt {
	return func(s *streamBetweenConfig) {
		s.name = name
	}
}

// StreamBetween allows a stream to traverse between two boundaries in a synchronized manner.
//
// Seems a bit strange to have this generated outside of the source boundary since the stream must be passed in.  In
// practice this should generally be invoked by the coordinating builder common between both sides.
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
			return err
		})
		return nil
	}

	//todo: feedback mechanism should be pluggable so we can avoid an extra goroutine
	go func() {
		ctx := context.Background()
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
