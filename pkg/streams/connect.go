package streams

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
)

type ConnectedPipe[E any] struct {
	from Source[E]
	sink Sink[E]

	data  *emitter.Subscription[E]
	drain *emitter.Subscription[Sink[E]]
}

func (c *ConnectedPipe[E]) Close(ctx context.Context) error {
	if err := c.from.Pause(ctx); err != nil {
		return err
	}
	c.from.SourceEvents().Data.Off(c.data)
	c.sink.SinkEvents().OnDrain.Off(c.drain)
	return nil
}

// Connect allows events to flow from a source to a sink.
func Connect[E any](ctx context.Context, from Source[E], to Sink[E]) (*ConnectedPipe[E], error) {
	sourceEvents := from.SourceEvents()
	sourceData := sourceEvents.Data.OnE(func(ctx context.Context, event E) error {
		return to.Write(ctx, event)
	})

	sinkEvents := to.SinkEvents()
	sinkDrain := sinkEvents.OnDrain.OnE(func(ctx context.Context, event Sink[E]) error {
		return from.Resume(ctx)
	})

	return &ConnectedPipe[E]{
		from:  from,
		sink:  to,
		data:  sourceData,
		drain: sinkDrain,
	}, from.Resume(ctx)
}
