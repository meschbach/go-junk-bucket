package emitter

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
)

// ListenerE is the typing for the function to receive events, possibly resulting in error on result
type ListenerE[E any] func(ctx context.Context, event E) error

// Listener is the typing for the function to receive events not resulting in an error
type Listener[E any] func(ctx context.Context, event E)

// Subscription represents a single listener bound to hear events
type Subscription[E any] struct {
	target ListenerE[E]
}

// Dispatcher manages a set of subscriptions and dispatching to those subscriptions.  Dispatcher is not thread or
// goproc safe.
type Dispatcher[E any] struct {
	listeners []*Subscription[E]
}

// OnE registers a given listener to receive events on the next broadcast.
func (e *Dispatcher[E]) OnE(l ListenerE[E]) *Subscription[E] {
	sub := &Subscription[E]{
		target: l,
	}
	e.listeners = append(e.listeners, sub)
	return sub
}

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
	dispatchTo := append(e.listeners)
	for _, l := range dispatchTo {
		if err := l.target(ctx, event); err != nil {
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}
