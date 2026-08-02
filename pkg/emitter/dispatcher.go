package emitter

import (
	"context"
	"errors"

	"github.com/meschbach/go-junk-bucket/pkg/fx"
)

// Dispatcher manages a set of subscriptions and dispatching to those
// subscriptions. Dispatcher is not safe for use by multiple goroutines; use
// MutexDispatcher when concurrent use is required.
type Dispatcher[E any] struct {
	listeners []*Subscription[E]
}

// NewDispatcher creates a new dispatcher for events of type E.
func NewDispatcher[E any]() *Dispatcher[E] {
	return &Dispatcher[E]{}
}

// OnE registers the given listener to receive events on subsequent broadcasts.
// A Subscription is returned to manage the listener; no de-duplication of
// listeners is performed.
func (e *Dispatcher[E]) OnE(l ListenerE[E]) *Subscription[E] {
	sub := &Subscription[E]{
		target: l,
		from:   e,
	}
	e.listeners = append(e.listeners, sub)
	return sub
}

// On registers the given listener to receive events on subsequent broadcasts.
func (e *Dispatcher[E]) On(l Listener[E]) *Subscription[E] {
	return e.OnE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Off removes the given subscription from further event broadcasts. It is
// idempotent.
func (e *Dispatcher[E]) Off(s *Subscription[E]) {
	e.listeners = fx.Filter(e.listeners, func(e *Subscription[E]) bool {
		return e != s
	})
}

// OnceE registers the given listener for a single broadcast; the subscription
// is removed from further broadcasts after the first delivery, even if the
// listener panics or returns an error.
func (e *Dispatcher[E]) OnceE(l ListenerE[E]) *Subscription[E] {
	var sub *Subscription[E]
	sub = e.OnE(func(ctx context.Context, event E) error {
		defer e.Off(sub)
		return l(ctx, event)
	})
	return sub
}

// Once registers the given listener for a single broadcast; the subscription
// is removed from further broadcasts after the first delivery.
func (e *Dispatcher[E]) Once(l Listener[E]) *Subscription[E] {
	return e.OnceE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Emit delivers the event to all registered listeners in registration order.
// A listener that returns an error, or panics, does not prevent the remaining
// listeners from receiving the event. When a single listener fails, its error
// is returned unchanged; when several fail, their errors are returned joined.
// nil is returned when no listener fails. Listeners added or removed during a
// dispatch take effect on the next Emit call.
func (e *Dispatcher[E]) Emit(ctx context.Context, event E) error {
	var problems []error
	dispatchTo := make([]*Subscription[E], len(e.listeners))
	copy(dispatchTo, e.listeners)
	for _, l := range dispatchTo {
		if err := invokeListener(ctx, l.target, event); err != nil {
			problems = append(problems, err)
		}
	}
	if len(problems) == 1 {
		return problems[0]
	}
	return errors.Join(problems...)
}
