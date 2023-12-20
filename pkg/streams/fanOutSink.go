package streams

import (
	"context"
	"errors"
)

func NewFanOutSink[T any]() *FanOutSink[T] {
	return &FanOutSink[T]{
		events: &SinkEvents[T]{},
	}
}

type FanOutSink[T any] struct {
	targets []Sink[T]
	events  *SinkEvents[T]
}

func (f *FanOutSink[T]) Add(target Sink[T]) {
	f.targets = append(f.targets, target)
}

func (f *FanOutSink[T]) Write(ctx context.Context, v T) error {
	var errs []error
	for _, target := range f.targets {
		if err := target.Write(ctx, v); err != nil {
			errs = append(errs, err)
		}
	}
	if errs == nil {
		return errors.Join(errs...)
	}
	return nil
}

func (f *FanOutSink[T]) Finish(ctx context.Context) error {
	var errs []error
	for _, target := range f.targets {
		if err := target.Finish(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if errs == nil {
		return errors.Join(errs...)
	}
	return nil
}

func (f *FanOutSink[T]) SinkEvents() *SinkEvents[T] {
	return f.events
}

func (f *FanOutSink[T]) Resume(ctx context.Context) error {
	return f.events.Full.Emit(ctx, f)
}
