package reactive

import (
	"context"
)

type ConnectedPipe[E any] struct {
	from Source[E]
	sink Sink[E]

	data  *Subscription[E]
	drain *Subscription[Sink[E]]
}

func (c *ConnectedPipe[E]) Close() error {
	c.from.SourceEvents().Data.Off(c.data)
	c.sink.SinkEvents().OnDrain.Off(c.drain)
	return nil
}

// Connect allows events to flow from a source to a sink.
func Connect[E any](ctx context.Context, from Source[E], to Sink[E]) (*ConnectedPipe[E], error) {
	sourceEvents := from.SourceEvents()
	sourceData := sourceEvents.Data.On(func(ctx context.Context, event E) error {
		return to.Write(ctx, event)
	})

	sinkEvents := to.SinkEvents()
	sinkDrain := sinkEvents.OnDrain.On(func(ctx context.Context, event Sink[E]) error {
		return from.Resume(ctx)
	})

	return &ConnectedPipe[E]{
		from:  from,
		sink:  to,
		data:  sourceData,
		drain: sinkDrain,
	}, from.Resume(ctx)
}
