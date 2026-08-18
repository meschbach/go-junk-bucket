package reactors

import "context"

// TickEventFunc is a handler invoked during a reactor tick.
// It receives the reactor context but no state; use [TickEventStateFunc] when
// state access is needed.
type TickEventFunc func(ctx context.Context) error

// TickEventStateFunc is a handler invoked during a reactor tick with access
// to the reactor's state S.
type TickEventStateFunc[S any] func(ctx context.Context, state S) error

// Boundary schedules work for execution within a single-threaded reactor domain.
// All scheduled functions execute sequentially, providing safe access to shared
// state without locks.
type Boundary[S any] interface {
	// ScheduleFunc queues operation for execution on the next reactor tick.
	ScheduleFunc(ctx context.Context, operation TickEventFunc)

	// ScheduleStateFunc queues operation for execution on the next reactor tick
	// with access to the reactor's state.
	ScheduleStateFunc(ctx context.Context, operation TickEventStateFunc[S])
}

// InvokeOp executes op within the context of reactor, storing the reactor
// boundary in the context passed to op.
func InvokeOp(underlying context.Context, reactor Boundary[any], op TickEventFunc) error {
	ctx := WithReactor(underlying, reactor)
	return op(ctx)
}

// InvokeStateOp executes op with the given state within the context of reactor,
// storing the reactor boundary in the context passed to op.
func InvokeStateOp[S any](underlying context.Context, reactor Boundary[S], state S, op TickEventStateFunc[S]) error {
	ctx := WithReactor[S](underlying, reactor)
	return op(ctx, state)
}
