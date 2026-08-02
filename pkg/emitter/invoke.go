package emitter

import (
	"context"
	"fmt"
)

// invokeListener calls target with the given context and event, converting any
// panic into an error. A single misbehaving listener therefore cannot abort
// delivery to the remaining listeners. An error used as a panic value is
// returned unchanged; any other value is coerced into an error describing the
// panic.
func invokeListener[E any](ctx context.Context, target ListenerE[E], event E) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("listener panic: %v", r)
			}
		}
	}()
	return target(ctx, event)
}
