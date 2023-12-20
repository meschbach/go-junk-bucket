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

	asyncTask := &task.Promise[R]{}
	target.ScheduleStateFunc(ctx, func(parentCtx context.Context, state O) error {
		VerifyWithinBoundary(parentCtx, target)
		ctx := parentCtx
		output, problem := apply(ctx, state)
		if problem != nil {
			span := trace.SpanFromContext(ctx)
			span.SetStatus(codes.Error, problem.Error())
		}

		replyTo.ScheduleFunc(ctx, func(ctx context.Context) error {
			VerifyWithinBoundary(ctx,replyTo)
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
