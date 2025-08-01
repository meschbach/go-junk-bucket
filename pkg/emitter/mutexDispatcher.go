package emitter

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
	"sync"
)

type MutexDispatcher[E any] struct {
	state     sync.Mutex
	listeners []*Subscription[E]
}

func NewMutexDispatcher[E any]() *MutexDispatcher[E] {
	return &MutexDispatcher[E]{
		state: sync.Mutex{},
	}
}

// OnE registers a given listener to receive events on the subsequent broadcasts.  A Subscription is provided to manage
// the invocation.  No de-duplication of listeners is performed.
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

// On registers a given listener to receive events on the next broadcast.
func (e *MutexDispatcher[E]) On(l Listener[E]) *Subscription[E] {
	return e.OnE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Off removes the given subscription from further event broadcasts.
func (e *MutexDispatcher[E]) Off(s *Subscription[E]) {
	e.state.Lock()
	defer e.state.Unlock()

	e.listeners = fx.Filter(e.listeners, func(e *Subscription[E]) bool {
		return e == s
	})
}

// OnceE registers the given listener l for a single broadcast then the subscription is removed from further broadcasts.
func (e *MutexDispatcher[E]) OnceE(l ListenerE[E]) *Subscription[E] {
	var sub *Subscription[E]
	sub = e.OnE(func(ctx context.Context, event E) error {
		err := l(ctx, event)
		e.Off(sub)
		return err
	})
	return sub
}

func (e *MutexDispatcher[E]) Once(l Listener[E]) *Subscription[E] {
	return e.OnceE(func(ctx context.Context, event E) error {
		l(ctx, event)
		return nil
	})
}

// Emit dispatches the event to all subscriptions returning a set of errors if any occur
func (e *MutexDispatcher[E]) Emit(ctx context.Context, event E) error {
	var dispatchTo []*Subscription[E]
	(func() {
		e.state.Lock()
		defer e.state.Unlock()

		//shallow clone of the slice
		dispatchTo = make([]*Subscription[E], len(e.listeners))
		copy(dispatchTo, e.listeners)
	})()

	var problems []error
	for _, l := range dispatchTo {
		if err := l.target(ctx, event); err != nil {
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}
