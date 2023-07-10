package reactors

import "context"

// WithReactor creates a child context with the specified reactor for contextual use
func WithReactor[S any](underlying context.Context, reactor Boundary[S]) context.Context {
	return context.WithValue(underlying, ContextKey, reactor)
}

// For retrieves a reactor Boundary from the given context
func For[S any](ctx context.Context) Boundary[S] {
	return ctx.Value(ContextKey).(Boundary[S])
}

func Maybe[S any](ctx context.Context) (boundary Boundary[S], has bool) {
	b, ok := ctx.Value(ContextKey).(Boundary[S])
	if !ok || b == nil {
		return nil, false
	}
	return b, true
}

// ScheduleFunc schedules an op on the contextual reactor Boundary.
func ScheduleFunc(ctx context.Context, op TickEventFunc) {
	For[any](ctx).ScheduleFunc(ctx, op)
}

const ContextKey = "meschbach.junk.reactor"
