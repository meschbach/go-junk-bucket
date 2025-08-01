package emitter

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
)

// Dispatcher manages a set of subscriptions and dispatching to those subscriptions.  Dispatcher is not safe to use
// with multiple goprocs.
type Dispatcher[E any] struct {
	listeners []*Subscription[E]
}

// NewDispatcher creates a new dispatcher for types of event E
func NewDispatcher[E any]() *Dispatcher[E] {
	return &Dispatcher[E]{}
}

// OnE registers a given listener to receive events on the subsequent broadcasts.  A Subscription is provided to manage
// the invocation.  No de-duplication of listeners is performed.
func (e *Dispatcher[E]) OnE(l ListenerE[E]) *Subscription[E] {
	sub := &Subscription[E]{
		target: l,
		from:   e,
	}
	e.listeners = append(e.listeners, sub)
	return sub
}

// On registers a given listener to receive events on the next broadcast.
func (e *Dispatcher[E]) On(l Listener[E]) *Subscription[E] {
	return e.OnE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Off removes the given subscription from further event broadcasts.
func (e *Dispatcher[E]) Off(s *Subscription[E]) {
	e.listeners = fx.Filter(e.listeners, func(e *Subscription[E]) bool {
		return e == s
	})
}

// OnceE registers the given listener l for a single broadcast then the subscription is removed from further broadcasts.
func (e *Dispatcher[E]) OnceE(l ListenerE[E]) *Subscription[E] {
	var sub *Subscription[E]
	sub = e.OnE(func(ctx context.Context, event E) error {
		err := l(ctx, event)
		e.Off(sub)
		return err
	})
	return sub
}

func (e *Dispatcher[E]) Once(l Listener[E]) *Subscription[E] {
	return e.OnceE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Emit dispatches the event to all subscriptions returning a set of errors if any occur
func (e *Dispatcher[E]) Emit(ctx context.Context, event E) error {
	var problems []error
	dispatchTo := make([]*Subscription[E], len(e.listeners))
	copy(dispatchTo, e.listeners)
	for _, l := range dispatchTo {
		if err := l.target(ctx, event); err != nil {
			problems = append(problems, err)
		}
	}
	if len(problems) == 1 {
		return problems[0]
	} else {
		return errors.Join(problems...)
	}
}
