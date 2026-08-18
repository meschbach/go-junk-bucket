package reactors

import (
	"context"
	"sync"
)

// Ticked is a [Boundary] that queues scheduled events for manual execution via [Ticked.Tick].
// Unlike [Channel], it has no internal goroutine and is driven entirely by the caller.
type Ticked[S any] struct {
	lock      sync.Mutex
	scheduled []TickEventStateFunc[S]
}

// Tick executes up to maximum pending events.
// Returns hasMore indicating whether events remain in the queue, and any error
// from a scheduled function.
func (t *Ticked[S]) Tick(ctx context.Context, maximum int, state S) (hasMore bool, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	completedTicks := 0
	for completedTicks < maximum {
		if len(t.scheduled) == 0 {
			select {
			case <-ctx.Done():
				return len(t.scheduled) > 0, ctx.Err()
			default:
				return false, nil
			}
		}
		next := t.scheduled[0]
		t.scheduled = t.scheduled[1:]

		if err := InvokeStateOp[S](ctx, t, state, next); err != nil {
			return len(t.scheduled) > 0, err
		}

		select {
		case <-ctx.Done():
			return len(t.scheduled) > 0, ctx.Err()
		default:
		}
	}
	return len(t.scheduled) > 0, nil
}

// ScheduleFunc appends operation to the pending queue.
// It will execute on the next call to [Ticked.Tick].
func (t *Ticked[S]) ScheduleFunc(ctx context.Context, operation TickEventFunc) {
	t.ScheduleStateFunc(ctx, func(ctx context.Context, state S) error {
		return operation(ctx)
	})
}

// ScheduleStateFunc appends operation to the pending queue.
// It will execute on the next call to [Ticked.Tick].
func (t *Ticked[S]) ScheduleStateFunc(ctx context.Context, operation TickEventStateFunc[S]) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.scheduled = append(t.scheduled, operation)
}
