package reactors

import "context"

// WithReactor returns a child context containing the given reactor boundary.
// The boundary can be retrieved later with [For] or [Maybe].
func WithReactor[S any](underlying context.Context, reactor Boundary[S]) context.Context {
	return context.WithValue(underlying, ContextKey, reactor)
}

// For returns the reactor [Boundary] stored in ctx.
// It panics if ctx is nil, contains no boundary, or the boundary type does not match S.
func For[S any](ctx context.Context) Boundary[S] {
	return ctx.Value(ContextKey).(Boundary[S])
}

// Maybe returns the reactor [Boundary] stored in ctx if present and type-compatible with S.
// Returns nil, false if the boundary is absent, ctx is nil, or the type does not match.
func Maybe[S any](ctx context.Context) (boundary Boundary[S], has bool) {
	if ctx == nil {
		return nil, false
	}
	b, ok := ctx.Value(ContextKey).(Boundary[S])
	if !ok || b == nil {
		return nil, false
	}
	return b, true
}

// ScheduleFunc schedules op on the reactor [Boundary] stored in ctx.
// Panics if ctx contains no boundary.
func ScheduleFunc(ctx context.Context, op TickEventFunc) {
	For[any](ctx).ScheduleFunc(ctx, op)
}

// ContextKey is the context value key used to store reactor boundaries.
const ContextKey = "meschbach.junk.reactor"
