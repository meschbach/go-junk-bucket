package task

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
)

func Map[I any, O any](ctx context.Context, from *Promise[I], mapFn func(ctx context.Context, input I) (O, error)) *Promise[O] {
	output := &Promise[O]{}
	from.OnCompleted(ctx, func(ctx context.Context, event Result[I]) {
		switch event.State {
		case Completed:
			nextValue, err := mapFn(ctx, event.Output)
			if err == nil {
				output.Success(ctx, nextValue)
			} else {
				output.Failure(ctx, err)
			}
		case Error:
			output.Failure(ctx, event.Problem)
		}
	})
	return output
}

// todo: reivew and merge with reactors.future.Promise
type Promise[O any] struct {
	result      Result[O]
	onCompleted emitter.Dispatcher[Result[O]]
}

func (p *Promise[O]) Then(ctx context.Context, handler emitter.Listener[O]) {
	p.OnCompleted(ctx, func(ctx context.Context, event Result[O]) {
		switch event.State {
		case Completed:
			handler(ctx, event.Output)
		}
	})
}

func (p *Promise[O]) OnCompleted(ctx context.Context, handler emitter.Listener[Result[O]]) {
	if p.result.Done() {
		handler(ctx, p.result)
	} else {
		p.onCompleted.On(handler)
	}
}

func (p *Promise[O]) Success(ctx context.Context, value O) {
	if p.result.State != Incomplete {
		panic("promise resolution called twice, success this time")
	}
	p.result.State = Completed
	p.result.Output = value

	p.dispatchResult(ctx)
}

func (p *Promise[O]) Failure(ctx context.Context, problem error) {
	if p.result.State != Incomplete {
		panic("promise resolution called twice, success this time")
	}

	p.result.State = Error
	p.result.Problem = problem

	p.dispatchResult(ctx)
}

func (p *Promise[O]) dispatchResult(ctx context.Context) {
	err := p.onCompleted.Emit(ctx, p.result)
	if err != nil {
		panic(err)
	}
}
