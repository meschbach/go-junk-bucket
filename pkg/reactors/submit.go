package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/task"
)

func Submit[I any, O any, R any](ctx context.Context, replyTo Boundary[I], target Boundary[O], apply func(boundaryContext context.Context, state O) (R, error)) *task.Promise[R] {
	asyncTask := &task.Promise[R]{}
	target.ScheduleStateFunc(ctx, func(ctx context.Context, state O) error {
		output, problem := apply(ctx, state)
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
