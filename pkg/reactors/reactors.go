// Package reactors provides elements to work with event reactors of various styles.  An event reactor is a
// de-multiplexer ensuring single threaded semantics within the domain of reactor.  These are useful in complex
// interactions between asynchronous components and simplifies message passing.
package reactors

import "context"

// TickEventFunc is a handler for a tick event within a reactor.  Provides no reactor state, so it must be passed in or
// extracted via other means
type TickEventFunc func(ctx context.Context) error

// TickEventStateFunc handles a tick event within a reactor given the reactor state S.
type TickEventStateFunc[S any] func(ctx context.Context, state S) error

// Boundary will de-multiplex multiple execution requests into a single serialized stream.
type Boundary[S any] interface {
	// ScheduleFunc schedules the given operation to be executed within the reactor upon the next tick of the given
	// reactor.
	ScheduleFunc(ctx context.Context, operation TickEventFunc)
	ScheduleStateFunc(ctx context.Context, operation TickEventStateFunc[S])
}

// InvokeOp invokes a given op within the context of the reactor.
func InvokeOp(underlying context.Context, reactor Boundary[any], op TickEventFunc) error {
	ctx := WithReactor(underlying, reactor)
	return op(ctx)
}

func InvokeStateOp[S any](underlying context.Context, reactor Boundary[S], state S, op TickEventStateFunc[S]) error {
	ctx := WithReactor[S](underlying, reactor)
	return op(ctx, state)
}
