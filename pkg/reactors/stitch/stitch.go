// Package stitch wraps suture to provide a channel based futures
package stitch

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors/futures"
)

type InitState[T any] func(ctx context.Context) (T, error)

type options struct {
	queueSize int
}

func New[T any](init InitState[T]) (*Actor[T], reactors.Boundary[T]) {
	opts := &options{queueSize: 32}
	processor, queue := reactors.NewChannel[T](opts.queueSize)
	return &Actor[T]{
		processor: processor,
		queue:     queue,
		init:      init,
	}, processor
}

type Actor[T any] struct {
	processor *reactors.Channel[T]
	queue     <-chan reactors.ChannelEvent[T]
	init      InitState[T]
}

func (a *Actor[T]) Serve(ctx context.Context) error {
	state, err := a.init(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case e := <-a.queue:
			if err := a.processor.Tick(ctx, e, state); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type ActorState[S any] struct {
	initialized bool
	value       S
}

// ConsumeAll processes all available work items within the work queue.
func (a *Actor[T]) ConsumeAll(ctx context.Context, state *ActorState[T]) (int, error) {
	consumedCount := 0
	if !state.initialized {
		if initState, err := a.init(ctx); err == nil {
			state.value = initState
			state.initialized = true
		} else {
			return consumedCount, err
		}
	}

	for {
		select {
		case e := <-a.queue:
			consumedCount++
			if err := a.processor.Tick(ctx, e, state.value); err != nil {
				return consumedCount, err
			}
		case <-ctx.Done():
			return consumedCount, ctx.Err()
		default:
			return consumedCount, nil
		}
	}
}

func (a *Actor[T]) Submit(ctx context.Context, fn func(context context.Context, state T) error) error {
	a.processor.ScheduleStateFunc(ctx, fn)
	return nil
}

func Promise[S any, T any](ctx context.Context, a *Actor[T], fn func(ctx context.Context, state T) (S, error)) *futures.Promise[T, S] {
	return futures.PromiseFuncOn[T, S](ctx, a.processor, fn)
}
