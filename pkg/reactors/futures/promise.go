package futures

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
)

type resolvedActions[O any] func(ctx context.Context, result Result[O])

// Promise represents a future computation running inside a reactor boundary.
//
// The type parameter S is the reactor state type the promise executes within,
// and O is the output type produced when the promise resolves.
//
// Results are delivered to handlers registered via [Promise.HandleFuncOn] or
// [Traverse], each executing within their own reactor boundary.
type Promise[S any, O any] struct {
	future  *Result[O]
	on      reactors.Boundary[S]
	pending []resolvedActions[O]
}

// PromiseFuncOn schedules op on reactor and returns a [Promise] that resolves
// when op completes.
//
// The op function executes inside the reactor boundary, with access to the
// reactor's state S. The returned Promise can be used to register handlers
// that receive the result on a different reactor boundary.
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

// HandleFuncOn registers op to run on the to reactor when this promise resolves.
//
// If the promise has already resolved, op is scheduled immediately. Otherwise,
// it is queued and scheduled when the promise completes. In both cases, op
// executes inside the to reactor boundary with access to its state S.
func (p *Promise[S, O]) HandleFuncOn(ctx context.Context, to reactors.Boundary[S], op func(ctx context.Context, state S, resolved Result[O]) error) {
	Traverse[S, O, S](ctx, p, to, op)
}

// Traverse routes a promise's resolution from its source reactor to a different
// target reactor.
//
// When the source promise resolves, op is scheduled on the to reactor boundary.
// If the promise is already resolved at the time Traverse is called, op is
// scheduled immediately. The type parameter T is the state type of the target
// reactor.
//
// Traverse is the underlying mechanism used by [Promise.HandleFuncOn]. Use it
// directly when the handler's reactor state type differs from the promise's
// source state type.
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

// Await blocks until this promise resolves or ctx is canceled.
//
// Await creates a temporary reactor internally to receive the resolution event.
// It should be used sparingly — prefer [Promise.HandleFuncOn] for non-blocking
// result handling within a reactor boundary.
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
