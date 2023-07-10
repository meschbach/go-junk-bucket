package futures

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
)

type resolvedActions[O any] func(ctx context.Context, result Result[O])

type Promise[S any, O any] struct {
	future  *Result[O]
	on      reactors.Boundary[S]
	pending []resolvedActions[O]
}

func PromiseFuncOn[S any, O any](ctx context.Context, reactor reactors.Boundary[S], op func(ctx context.Context, state S) (O, error)) *Promise[S, O] {
	promised := &Promise[S, O]{
		future: &Result[O]{
			Resolved: false,
		},
		on: reactor,
	}
	reactor.ScheduleStateFunc(ctx, func(ctx context.Context, state S) error {
		value, err := op(ctx, state)
		promised.future.Resolved = true
		promised.future.Result = value
		promised.future.Problem = err

		for _, resolvers := range promised.pending {
			resolvers(ctx, *promised.future)
		}

		return nil
	})
	return promised
}

func (p *Promise[S, O]) HandleFuncOn(ctx context.Context, to reactors.Boundary[S], op func(ctx context.Context, state S, resolved Result[O]) error) {
	Traverse[S, O, S](ctx, p, to, op)
}

func Traverse[S any, O any, T any](ctx context.Context, source *Promise[S, O], to reactors.Boundary[T], op func(ctx context.Context, state T, resolved Result[O]) error) {
	source.on.ScheduleStateFunc(ctx, func(ctx context.Context, state S) error {
		if source.future.Resolved {
			to.ScheduleStateFunc(ctx, func(ctx context.Context, state T) error {
				return op(ctx, state, *source.future)
			})
		} else {
			source.pending = append(source.pending, func(ctx context.Context, result Result[O]) {
				to.ScheduleStateFunc(ctx, func(ctx context.Context, state T) error {
					return op(ctx, state, result)
				})
			})
		}
		return nil
	})
}

type awaitReactorState struct{}

// Await will wait for completion of a promise on a temporarily created reactor.
func (p *Promise[S, O]) Await(ctx context.Context) (Result[O], error) {
	state := &awaitReactorState{}
	demultiplexer, input := reactors.NewChannel[*awaitReactorState](1)
	defer demultiplexer.Done()
	var out Result[O]
	Traverse[S, O, *awaitReactorState](ctx, p, demultiplexer, func(ctx context.Context, s *awaitReactorState, resolved Result[O]) error {
		out = resolved
		return nil
	})
	var err error
	select {
	case e := <-input:
		err = demultiplexer.Tick(ctx, e, state)
	case <-ctx.Done():
		err = ctx.Err()
	}
	return out, err
}
