package emitter

import (
	"context"
	"errors"
	"sync"

	"github.com/meschbach/go-junk-bucket/pkg/fx"
)

// MutexDispatcher manages a set of subscriptions and dispatching to those
// subscriptions. It is safe for concurrent use by multiple goroutines:
// registration and removal are serialized by a mutex, and events are
// dispatched to a snapshot of the listeners so a listener may safely register
// or unsubscribe while a dispatch is in progress.
type MutexDispatcher[E any] struct {
	state     sync.Mutex
	listeners []*Subscription[E]
}

// NewMutexDispatcher creates a new concurrency-safe dispatcher for events of
// type E.
func NewMutexDispatcher[E any]() *MutexDispatcher[E] {
	return &MutexDispatcher[E]{
		state: sync.Mutex{},
	}
}

// OnE registers the given listener to receive events on subsequent broadcasts.
// A Subscription is returned to manage the listener; no de-duplication of
// listeners is performed.
func (e *MutexDispatcher[E]) OnE(l ListenerE[E]) *Subscription[E] {
	sub := &Subscription[E]{
		target: l,
		from:   e,
	}

	e.state.Lock()
	defer e.state.Unlock()
	e.listeners = append(e.listeners, sub)
	return sub
}

// On registers the given listener to receive events on subsequent broadcasts.
func (e *MutexDispatcher[E]) On(l Listener[E]) *Subscription[E] {
	return e.OnE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Off removes the given subscription from further event broadcasts. It is
// idempotent.
func (e *MutexDispatcher[E]) Off(s *Subscription[E]) {
	e.state.Lock()
	defer e.state.Unlock()

	e.listeners = fx.Filter(e.listeners, func(e *Subscription[E]) bool {
		return e != s
	})
}

// OnceE registers the given listener for a single broadcast; the subscription
// is removed from further broadcasts after the first delivery, even if the
// listener panics or returns an error.
func (e *MutexDispatcher[E]) OnceE(l ListenerE[E]) *Subscription[E] {
	var sub *Subscription[E]
	sub = e.OnE(func(ctx context.Context, event E) error {
		defer e.Off(sub)
		return l(ctx, event)
	})
	return sub
}

// Once registers the given listener for a single broadcast; the subscription
// is removed from further broadcasts after the first delivery.
func (e *MutexDispatcher[E]) Once(l Listener[E]) *Subscription[E] {
	return e.OnceE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Emit delivers the event to all registered listeners in registration order.
// A listener that returns an error, or panics, does not prevent the remaining
// listeners from receiving the event. When a single listener fails, its error
// is returned unchanged; when several fail, their errors are returned joined.
// nil is returned when no listener fails. Listeners are invoked after the
// mutex is released, so a listener may safely register or unsubscribe during a
// dispatch; changes take effect on the next Emit call.
func (e *MutexDispatcher[E]) Emit(ctx context.Context, event E) error {
	e.state.Lock()
	dispatchTo := make([]*Subscription[E], len(e.listeners))
	copy(dispatchTo, e.listeners)
	e.state.Unlock()

	var problems []error
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
