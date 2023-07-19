package reactive

import (
	"context"
	"errors"
)

type TransformerState uint8

const (
	TransformerPaused TransformerState = iota
	TransformerFlowing
	TransformerFinished
)

type TransformerFunc[I any, O any] func(ctx context.Context, input I) (output O, err error)

type Transformer[I any, O any] struct {
	sourceEvents *SourceEvents[O]
	sinkEvents   *SinkEvents[I]
	fn           TransformerFunc[I, O]
	state        TransformerState
	buffer       []O
	maximum      int
}

func NewTransform[I any, O any](fn TransformerFunc[I, O]) *Transformer[I, O] {
	return &Transformer[I, O]{
		sourceEvents: &SourceEvents[O]{},
		sinkEvents:   &SinkEvents[I]{},
		fn:           fn,
		maximum:      10,
	}
}

func (t *Transformer[I, O]) Pump(ctx context.Context, source Source[I], target Sink[O]) error {
	target.SinkEvents().OnDrain.On(func(ctx context.Context, event Sink[O]) error {
		return t.Resume(ctx)
	})
	t.sinkEvents.OnDrain.On(func(ctx context.Context, event Sink[I]) error {
		return source.Resume(ctx)
	})
	t.sourceEvents.Data.On(func(ctx context.Context, event O) error {
		return target.Write(ctx, event)
	})
	source.SourceEvents().Data.On(func(ctx context.Context, event I) error {
		return t.Write(ctx, event)
	})
	return target.Resume(ctx)
}

func (t *Transformer[I, O]) Resume(ctx context.Context) error {
	switch t.state {
	case TransformerFinished:
		return Done
	}
	t.state = TransformerFlowing
	if err := t.pumpInternal(ctx); err != nil {
		return err
	}
	return t.sinkEvents.OnDrain.Emit(ctx, t)
}

func (t *Transformer[I, O]) pumpInternal(ctx context.Context) error {
	for i, e := range t.buffer {
		err := t.sourceEvents.Data.Emit(ctx, e)

		if err == nil {
			continue
		}
		//remove pushed elements
		if i > 0 { // not first
			t.buffer = t.buffer[i-1:]
		}
		//todo: state cycles?
		return err
	}
	t.buffer = nil
	return nil
}

func (t *Transformer[I, O]) Write(ctx context.Context, v I) error {
	switch t.state {
	case TransformerPaused:
		if len(t.buffer) >= t.maximum {
			return Full
		}
	case TransformerFinished:
		return Done
	}

	out, err := t.fn(ctx, v)
	if err != nil {
		return err
	}

	switch t.state {
	case TransformerPaused:
		t.buffer = append(t.buffer, out)
	case TransformerFlowing:
		if t.buffer != nil {
			t.buffer = append(t.buffer, out)
			return t.pumpInternal(ctx)
		} else {
			problem := t.sourceEvents.Data.Emit(ctx, out)
			switch problem {
			case Full:
				t.buffer = append(t.buffer, out)
			}
			return problem
		}
	}
	return nil
}

func (t *Transformer[I, O]) Finish(ctx context.Context) error {
	if t.state == TransformerFinished {
		return nil
	}
	t.state = TransformerFinished
	//todo: should only be emitted after it is completed
	return t.sinkEvents.OnFinished.Emit(ctx, t)
}

func (t *Transformer[I, O]) SinkEvents() *SinkEvents[I] {
	return t.sinkEvents
}

func (t *Transformer[I, O]) SourceEvents() *SourceEvents[O] {
	return t.sourceEvents
}

func (t *Transformer[I, O]) ReadSlice(ctx context.Context, to []O) (int, error) {
	return 0, errors.New("todo")
}
