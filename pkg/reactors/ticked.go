package reactors

import "context"

// Ticked is a Reactor externally driven when calling the Tick method.  Events will be queued until it is
// manually ticked.
type Ticked struct {
	scheduled []TickEventFunc
}

// Tick executes up to the maximum number of event reductions within the reactor.
func (t *Ticked) Tick(ctx context.Context, maximum int) (hasMore bool, err error) {
	completedTicks := 0
	for completedTicks < maximum {
		if len(t.scheduled) == 0 {
			return false, nil
		}
		next := t.scheduled[0]
		t.scheduled = t.scheduled[1:]

		if err := InvokeOp(ctx, t, next); err != nil {
			return len(t.scheduled) > 0, err
		}
	}
	return len(t.scheduled) > 0, nil
}

func (t *Ticked) ScheduleFunc(ctx context.Context, operation TickEventFunc) {
	t.scheduled = append(t.scheduled, operation)
}
