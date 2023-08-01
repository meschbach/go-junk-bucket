package emitter

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
)

// Listener is the typing for the function to receive events
type Listener[E any] func(ctx context.Context, event E) error

// Subscription represents a single listener bound to hear events
type Subscription[E any] struct {
	//tag represents the unique ID of the event emitter
	tag    int
	target Listener[E]
}

// EventEmitter manages a set of subscriptions and dispatching to those subscriptions.  EventEmitter is not thread or
// goproc safe.
type EventEmitter[E any] struct {
	//nextTag is the unique tag value to issue to the next subscriber
	nextTag   int
	listeners []*Subscription[E]
}

// On registers a given listener to receive events on the next broadcast.
func (e *EventEmitter[E]) On(l Listener[E]) *Subscription[E] {
	tag := e.nextTag
	e.nextTag++
	sub := &Subscription[E]{
		tag:    tag,
		target: l,
	}
	e.listeners = append(e.listeners, sub)
	return sub
}

// Off removes the given subscription from further event broadcasts.
func (e *EventEmitter[E]) Off(s *Subscription[E]) {
	e.listeners = fx.Filter(e.listeners, func(e *Subscription[E]) bool {
		return e.tag != s.tag
	})
}

// Once registers the given listener l for a single broadcast then the subscription is removed from further broadcasts.
func (e *EventEmitter[E]) Once(l Listener[E]) *Subscription[E] {
	var sub *Subscription[E]
	sub = e.On(func(ctx context.Context, event E) error {
		err := l(ctx, event)
		e.Off(sub)
		return err
	})
	return sub
}

// Emit dispatches the event to all subscriptions returning a set of errors if any occur
func (e *EventEmitter[E]) Emit(ctx context.Context, event E) error {
	var problems []error
	dispatchTo := append(e.listeners)
	for _, l := range dispatchTo {
		if err := l.target(ctx, event); err != nil {
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}
