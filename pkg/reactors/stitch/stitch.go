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

func New[T any](init InitState[T]) *Actor[T] {
	opts := &options{queueSize: 32}
	processor, queue := reactors.NewChannel[T](opts.queueSize)
	return &Actor[T]{
		processor: processor,
		queue:     queue,
		init:      init,
	}
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

func Promise[S any, T any](ctx context.Context, a *Actor[T], fn func(ctx context.Context, state T) (S, error)) *futures.Promise[T, S] {
	return futures.PromiseFuncOn[T, S](ctx, a.processor, fn)
}
