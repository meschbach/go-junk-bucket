package reactors

import "context"

// WithReactor creates a child context with the specified reactor for contextual use
func WithReactor(underlying context.Context, reactor Boundary) context.Context {
	return context.WithValue(underlying, ContextKey, reactor)
}

// For retrieves a reactor Boundary from the given context
func For(ctx context.Context) Boundary {
	return ctx.Value(ContextKey).(Boundary)
}

// ScheduleFunc schedules an op on the contextual reactor Boundary.
func ScheduleFunc(ctx context.Context, op TickEventFunc) {
	For(ctx).ScheduleFunc(ctx, op)
}

const ContextKey = "meschbach.junk.reactor"
