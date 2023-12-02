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
	cfg := streamBetweenConfig{}
	cfg.init(opts)

	//todo: this has many edge cases which will be paid off over time.
	//arguably this procedure belongs in another package entirely since it is a union between reactors and streams
	inputSink := streams.NewBuffer[E](32, streams.WithBufferTracePrefix[E](cfg.name+".sink"))
	outputSource := streams.NewBuffer[E](32, streams.WithBufferTracePrefix[E](cfg.name+".source"))

	inputEvents := inputSink.SourceEvents()
	inputEvents.Data.On(func(inputContext context.Context, event E) {
		VerifyWithinBoundary(inputContext, inputSide)
		outputSide.ScheduleFunc(inputContext, func(outputContext context.Context) error {
			VerifyWithinBoundary(outputContext, outputSide)
			//todo: feedback and propagation of signals
			return outputSource.Write(outputContext, event)
		})
	})
	inputSink.SinkEvents().OnFinished.On(func(inputContext context.Context, s streams.Sink[E]) {
		VerifyWithinBoundary[I](inputContext, inputSide)
		outputSide.ScheduleFunc(inputContext, func(outputContext context.Context) error {
			VerifyWithinBoundary[O](outputContext, outputSide)
			//todo: feedback and propagation of signals
			return outputSource.Finish(outputContext)
		})
	})

	inputSide.ScheduleFunc(ctx, func(ctx context.Context) error {
		return inputSink.Resume(ctx)
	})

	return outputSource, inputSink, nil
}
