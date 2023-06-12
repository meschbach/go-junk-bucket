package reactors

import "context"

func WithReactor(underlying context.Context, reactor Reactor) context.Context {
	return context.WithValue(underlying, ContextKey, reactor)
}

func For(ctx context.Context) Reactor {
	return ctx.Value(ContextKey).(Reactor)
}

func ScheduleFunc(ctx context.Context, op TickEventFunc) {
	For(ctx).ScheduleFunc(ctx, op)
}

const ContextKey = "meschbach.junk.reactor"
