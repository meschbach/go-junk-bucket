package streams

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ConnectedPipeOption func(c *connectedPipeConfig) error

func WithTraceAttributes(attrs ...attribute.KeyValue) ConnectedPipeOption {
	return func(c *connectedPipeConfig) error {
		c.traceAttributes = attrs
		return nil
	}
}
func WithTracePrefix(tracePrefix string) ConnectedPipeOption {
	return func(c *connectedPipeConfig) error {
		c.traceEventPrefix = tracePrefix
		return nil
	}
}

func WithSuppressEnd() ConnectedPipeOption {
	return func(c *connectedPipeConfig) error {
		c.suppressEnd = true
		return nil
	}
}

type connectedPipeConfig struct {
	traceAttributes  []attribute.KeyValue
	traceEventPrefix string
	suppressEnd      bool
}

func (c *connectedPipeConfig) init(opts []ConnectedPipeOption) error {
	var errs []error
	for _, opt := range opts {
		errs = append(errs, opt(c))
	}
	if errs != nil {
		return errors.Join(errs...)
	}
	return nil
}

type ConnectedPipe[E any] struct {
	from Source[E]
	sink Sink[E]

	onSourceData     *emitter.Subscription[E]
	onSourceFinished *emitter.Subscription[Source[E]]
	full             *emitter.Subscription[Sink[E]]
	drain            *emitter.Subscription[Sink[E]]
	config           connectedPipeConfig
}

func (c *ConnectedPipe[E]) Close(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(c.config.traceEventPrefix+".Close", trace.WithAttributes(c.config.traceAttributes...))

	if err := c.from.Pause(ctx); err != nil {
		return err
	}
	sourceEvents := c.from.SourceEvents()
	sourceEvents.Data.Off(c.onSourceData)
	if !c.config.suppressEnd {
		sourceEvents.End.Off(c.onSourceFinished)
	}

	sinkEvents := c.sink.SinkEvents()
	sinkEvents.Drained.Off(c.drain)
	sinkEvents.Full.Off(c.full)
	return nil
}

// Connect allows events to flow from a source to a sink.
func Connect[E any](ctx context.Context, from Source[E], to Sink[E], opts ...ConnectedPipeOption) (*ConnectedPipe[E], error) {
	pipe := &ConnectedPipe[E]{
		from: from,
		sink: to,
		config: connectedPipeConfig{
			traceEventPrefix: "ConnectedPipe",
		},
	}
	//todo: should become a default opts
	if err := pipe.config.init(opts); err != nil {
		return nil, err
	}
	opts = nil

	attrs := trace.WithAttributes(pipe.config.traceAttributes...)
	span := trace.SpanFromContext(ctx)
	span.AddEvent(pipe.config.traceEventPrefix+".Connect", attrs)

	sourceEvents := from.SourceEvents()
	pipe.onSourceData = sourceEvents.Data.OnE(func(ctx context.Context, event E) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(pipe.config.traceEventPrefix+".Source.Data", attrs)
		writeErr := to.Write(ctx, event)
		return writeErr
	})
	if !pipe.config.suppressEnd {
		pipe.onSourceFinished = sourceEvents.End.OnE(func(ctx context.Context, event Source[E]) error {
			span := trace.SpanFromContext(ctx)
			span.AddEvent(pipe.config.traceEventPrefix+".Source.End", attrs)
			return to.Finish(ctx)
		})
	}

	sinkEvents := to.SinkEvents()
	pipe.full = sinkEvents.Full.OnE(func(ctx context.Context, event Sink[E]) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(pipe.config.traceEventPrefix+".Sink.Full", attrs)
		return from.Pause(ctx)
	})
	pipe.drain = sinkEvents.Drained.OnE(func(ctx context.Context, event Sink[E]) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(pipe.config.traceEventPrefix+".Sink.Drain", attrs)

		result := from.Resume(ctx)
		if pipe.config.suppressEnd && result == End {
			result = nil
		}
		return result
	})
	result := from.Resume(ctx)
	if result != nil {
		if result == Full || result == UnderRun {
			result = nil
		}
	}
	return pipe, result
}
