package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracing = otel.Tracer("github.com/meschbach/go-junk-bucket/pkg/reactors")

func Submit[I any, O any, R any](ctx context.Context, replyTo Boundary[I], target Boundary[O], apply func(boundaryContext context.Context, state O) (R, error)) *task.Promise[R] {
	VerifyWithinBoundary(ctx, replyTo)
	span := trace.SpanFromContext(ctx)
	callingContext := span.SpanContext()

	asyncTask := &task.Promise[R]{}
	target.ScheduleStateFunc(ctx, func(parentCtx context.Context, state O) error {
		VerifyWithinBoundary(parentCtx, target)
		ctx := trace.ContextWithRemoteSpanContext(parentCtx, callingContext)
		output, problem := apply(ctx, state)
		if problem != nil {
			span.SetStatus(codes.Error, problem.Error())
		}

		replyTo.ScheduleFunc(ctx, func(ctx context.Context) error {
			if problem == nil {
				asyncTask.Success(ctx, output)
			} else {
				asyncTask.Failure(ctx, problem)
			}
			return nil
		})
		return nil
	})
	return asyncTask
}
