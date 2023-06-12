// Package reactors provides elements to work with event reactors of various styles.  An event reactor is a
// de-multiplexer ensuring single threaded semantics within the domain of reactor.  These are useful in complex
// interactions between asynchronous components and simplifies message passing.
package reactors

import "context"

// TickEventFunc is a handler for a tick event within a reactor.
type TickEventFunc func(ctx context.Context) error

// Reactor will de-multiplex multiple execution requests into a single serialized stream.
type Reactor interface {
	// ScheduleFunc schedules the given operation to be executed within the reactor upon the next tick of the given
	// reactor.
	ScheduleFunc(ctx context.Context, operation TickEventFunc)
}

// InvokeOp invokes a given op within the context of the reactor.
func InvokeOp(underlying context.Context, reactor Reactor, op TickEventFunc) error {
	ctx := WithReactor(underlying, reactor)
	return op(ctx)
}
