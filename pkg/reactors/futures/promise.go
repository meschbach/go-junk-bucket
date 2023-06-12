package futures

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
)

type resolvedActions[O any] struct {
	tell reactors.Reactor
	op   func(ctx context.Context, resolved Result[O]) error
}

type Promise[O any] struct {
	future  *Result[O]
	on      reactors.Reactor
	pending []resolvedActions[O]
}

func PromiseFuncOn[O any](ctx context.Context, reactor reactors.Reactor, op func(ctx context.Context) (O, error)) *Promise[O] {
	promised := &Promise[O]{
		future: &Result[O]{
			Resolved: false,
		},
		on: reactor,
	}
	reactor.ScheduleFunc(ctx, func(ctx context.Context) error {
		value, err := op(ctx)
		promised.future.Resolved = true
		promised.future.Result = value
		promised.future.Problem = err

		for _, resolvers := range promised.pending {
			resolvers.tell.ScheduleFunc(ctx, func(ctx context.Context) error {
				return resolvers.op(ctx, *promised.future)
			})
		}

		return nil
	})
	return promised
}

func (p *Promise[O]) HandleFuncOn(ctx context.Context, to reactors.Reactor, op func(ctx context.Context, resolved Result[O]) error) {
	p.on.ScheduleFunc(ctx, func(ctx context.Context) error {
		if p.future.Resolved {
			to.ScheduleFunc(ctx, func(ctx context.Context) error {
				return op(ctx, *p.future)
			})
		} else {
			p.pending = append(p.pending, resolvedActions[O]{
				tell: to,
				op:   op,
			})
		}
		return nil
	})
}

// Await will wait for completion of a promise on a temporarily created reactor.
func (p *Promise[O]) Await(ctx context.Context) (Result[O], error) {
	demultiplexer, input := reactors.NewChannel(1)
	defer demultiplexer.Done()
	var out Result[O]
	p.HandleFuncOn(ctx, demultiplexer, func(ctx context.Context, resolved Result[O]) error {
		out = resolved
		return nil
	})
	var err error
	select {
	case e := <-input:
		err = demultiplexer.Tick(ctx, e)
	case <-ctx.Done():
		err = ctx.Err()
	}
	return out, err
}
