package stitch

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors/futures"
)

// InitState is a function that initializes an Actor's state.
// It is called lazily on first use (either Serve or ConsumeAll).
type InitState[T any] func(ctx context.Context) (T, error)

type options struct {
	queueSize int
}

// New creates an [Actor] with the given state initializer.
// It returns the Actor and its [reactors.Boundary] for scheduling work.
func New[T any](init InitState[T]) (*Actor[T], reactors.Boundary[T]) {
	opts := &options{queueSize: 32}
	processor, queue := reactors.NewChannel[T](opts.queueSize)
	return &Actor[T]{
		processor: processor,
		queue:     queue,
		init:      init,
	}, processor
}

// Actor is a self-contained reactor unit managing its own state lifecycle.
//
// An Actor runs a [reactors.Channel] event loop with lazy-initialized state.
// The state type T is the shared state accessible to all scheduled operations.
//
// Use [Actor.Serve] for goroutine-based operation (compatible with suture supervisor
// trees), or [Actor.ConsumeAll] for manual/immediate-mode driving.
type Actor[T any] struct {
	processor *reactors.Channel[T]
	queue     <-chan reactors.ChannelEvent[T]
	init      InitState[T]
}

// Serve runs the Actor's event loop, initializing state and processing events
// until ctx is done. The state is initialized lazily on the first event.
//
// Serve blocks in the calling goroutine and should not spawn additional
// goroutines. If the init function returns an error, Serve returns immediately.
//
// Serve satisfies the suture.Service interface, making it compatible with
// suture supervisor trees.
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

// ActorState wraps state for manual/immediate-mode driving with [Actor.ConsumeAll].
//
// Pass a zero-value ActorState to ConsumeAll on first call; it will initialize
// the state automatically. On subsequent calls, pass the same ActorState to
// reuse the initialized state.
type ActorState[S any] struct {
	initialized bool
	value       S
}

// ConsumeAll processes all available work items within the work queue.
//
// On first call, the Actor's state is initialized via [InitState] and stored
// in state. On subsequent calls, the existing state is reused.
//
// Returns the number of events consumed and any error from processing or
// context cancellation.
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

// Submit schedules fn for execution within the Actor's reactor boundary.
func (a *Actor[T]) Submit(ctx context.Context, fn func(context context.Context, state T) error) error {
	a.processor.ScheduleStateFunc(ctx, fn)
	return nil
}

// Promise creates a [futures.Promise] that runs fn within the Actor's reactor boundary.
//
// The fn function receives the Actor's state T and produces output of type S.
// The returned Promise resolves when fn completes, and can be used to deliver
// the result to a different reactor boundary via [futures.Promise.HandleFuncOn].
func Promise[S any, T any](ctx context.Context, a *Actor[T], fn func(ctx context.Context, state T) (S, error)) *futures.Promise[T, S] {
	return futures.PromiseFuncOn[T, S](ctx, a.processor, fn)
}
