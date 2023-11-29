package streams

import (
	"context"
	"errors"
)

type fixedSlice[T any] struct {
	events *SourceEvents[T]
	values []T
}

func (f *fixedSlice[T]) ReadSlice(ctx context.Context, to []T) (int, error) {
	if len(f.values) == 0 {
		return 0, End
	}
	copied := copy(to, f.values)
	f.values = f.values[copied:]
	return copied, nil
}

func (f *fixedSlice[T]) SourceEvents() *SourceEvents[T] {
	return f.events
}

func (f *fixedSlice[T]) Resume(ctx context.Context) error {
	b := make([]T, 1)
	for { //todo: only while flowing
		count, err := f.ReadSlice(ctx, b)
		if err != nil {
			return err
		}
		if count == 0 {
			//todo: should stop flowing
			return Done
		}
		switch err := f.events.Data.Emit(ctx, b[0]); err {
		case nil:
			continue
		default:
			return err
		}
	}
}

// todo: write implementation
func (f *fixedSlice[T]) Pause(ctx context.Context) error {
	return errors.New("todo")
}

func FromSlice[T any](values []T) Source[T] {
	return &fixedSlice[T]{
		events: &SourceEvents[T]{},
		values: values,
	}
}
