package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
)

// StreamBetween allows a stream to traverse between two boundaries in a synchronized manner.
//
// Seems a bit strange to have this generated outside of the source boundary since the stream must be passed in.  In
// practice this should generally be invoked by the coordinating builder common between both sides.
func StreamBetween[E any, I any, O any](ctx context.Context, source Boundary[I], target Boundary[O]) (streams.Source[E], streams.Sink[E], error) {
	//todo: this has many edge cases which will be paid off over time.
	//arguably this procedure belongs in another package entirely since it is a union between reactors and streams
	sinkSide := streams.NewBuffer[E](10)
	readSide := streams.NewBuffer[E](10)

	sinkSideEvents := sinkSide.SourceEvents()
	sinkSideEvents.Data.On(func(ctx context.Context, event E) {
		target.ScheduleFunc(ctx, func(ctx context.Context) error {
			//todo: feedback and propagation of signals
			readSide.Write(ctx, event)
			return nil
		})
	})
	sinkSide.SinkEvents().OnFinished.On(func(ctx context.Context, s streams.Sink[E]) {
		target.ScheduleFunc(ctx, func(ctx context.Context) error {
			//todo: feedback and propagation of signals
			readSide.Finish(ctx)
			return nil
		})
	})

	return readSide, sinkSide, sinkSide.Resume(ctx)
}
