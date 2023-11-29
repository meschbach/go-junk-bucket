package streams

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ConnectedPipe[E any] struct {
	from Source[E]
	sink Sink[E]

	onSourceData     *emitter.Subscription[E]
	onSourceFinished *emitter.Subscription[Source[E]]
	drain            *emitter.Subscription[Sink[E]]
}

func (c *ConnectedPipe[E]) Close(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("ConnectedPipe.Close")

	if err := c.from.Pause(ctx); err != nil {
		return err
	}
	sourceEvents := c.from.SourceEvents()
	sourceEvents.Data.Off(c.onSourceData)
	sourceEvents.End.Off(c.onSourceFinished)
	c.sink.SinkEvents().OnDrain.Off(c.drain)
	return nil
}

// Connect allows events to flow from a source to a sink.
func Connect[E any](ctx context.Context, from Source[E], to Sink[E]) (*ConnectedPipe[E], error) {
	pipe := &ConnectedPipe[E]{
		from: from,
		sink: to,
	}

	attrs := trace.WithAttributes(attribute.Stringer("ConnectedPipe", ptrFormattedStringer[ConnectedPipe[E]]{pipe}))
	span := trace.SpanFromContext(ctx)
	span.AddEvent("ConnectedPipe.Connect", attrs)

	sourceEvents := from.SourceEvents()
	pipe.onSourceData = sourceEvents.Data.OnE(func(ctx context.Context, event E) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent("ConnectedPipe.Source.Data", attrs)
		return to.Write(ctx, event)
	})
	pipe.onSourceFinished = sourceEvents.End.OnE(func(ctx context.Context, event Source[E]) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent("ConnectedPipe.Source.End", attrs)
		return to.Finish(ctx)
	})

	sinkEvents := to.SinkEvents()
	pipe.drain = sinkEvents.OnDrain.OnE(func(ctx context.Context, event Sink[E]) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent("ConnectedPipe.Source.Data", attrs)

		return from.Resume(ctx)
	})

	return pipe, from.Resume(ctx)
}
