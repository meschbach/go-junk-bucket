package reactors

import (
	"context"

	"github.com/meschbach/go-junk-bucket/pkg/task"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracing = otel.Tracer("github.com/meschbach/go-junk-bucket/pkg/reactors")

// Submit runs apply on the target reactor and returns a [task.Promise] that resolves
// on the replyTo reactor.
//
// Call Submit from within a reactor (replyTo) to request work from another reactor
// (target). The apply function executes inside target's boundary, and the result
// is delivered back to replyTo's boundary — all without leaving the single-threaded
// guarantees of either reactor.
//
// The replyTo boundary must match the boundary in ctx; [VerifyWithinBoundary] will
// panic if called with the sane build tag and the boundaries do not match.
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
			VerifyWithinBoundary(ctx, replyTo)
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
