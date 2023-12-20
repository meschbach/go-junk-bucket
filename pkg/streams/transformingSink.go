package streams

import "context"

type TransformFunc[I any, O any] func(ctx context.Context, in I) (O, error)

func WrapTransformingSink[I any, O any](wrap Sink[O], transformer TransformFunc[I, O]) *TransformingSink[I, O] {
	wrapper := &TransformingSink[I, O]{
		next:      wrap,
		transform: transformer,
		events:    &SinkEvents[I]{},
	}
	upstreamEvents := wrap.SinkEvents()
	upstreamEvents.Available.OnE(func(ctx context.Context, event SinkAvailableEvent[O]) error {
		return wrapper.events.Available.Emit(ctx, SinkAvailableEvent[I]{
			Space: event.Space,
			Sink:  wrapper,
		})
	})
	upstreamEvents.Drained.OnE(func(ctx context.Context, event Sink[O]) error {
		return wrapper.events.Drained.Emit(ctx, wrapper)
	})
	upstreamEvents.Full.OnE(func(ctx context.Context, event Sink[O]) error {
		return wrapper.events.Full.Emit(ctx, wrapper)
	})
	upstreamEvents.Finishing.OnE(func(ctx context.Context, event Sink[O]) error {
		return wrapper.events.Finishing.Emit(ctx, wrapper)
	})
	upstreamEvents.Finished.OnE(func(ctx context.Context, event Sink[O]) error {
		return wrapper.events.Finished.Emit(ctx, wrapper)
	})
	return wrapper
}

type TransformingSink[I any, O any] struct {
	next      Sink[O]
	transform TransformFunc[I, O]
	events    *SinkEvents[I]
}

func (t *TransformingSink[I, O]) Write(ctx context.Context, in I) error {
	transformed, err := t.transform(ctx, in)
	if err != nil {
		return err
	}
	return t.next.Write(ctx, transformed)
}

func (t *TransformingSink[I, O]) Finish(ctx context.Context) error {
	return t.next.Finish(ctx)
}

func (t *TransformingSink[I, O]) SinkEvents() *SinkEvents[I] {
	return t.events
}

func (t *TransformingSink[I, O]) Resume(ctx context.Context) error {
	return t.next.Resume(ctx)
}
