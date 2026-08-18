/*
Package emitter provides 1:n event dispatch, the publish/subscribe pattern in
which a single Emit delivers one event to every registered listener.

# Motivation

Direct wiring between a producer and its consumers couples their lifecycles:
each must know the other exists, and adding or removing a consumer means
touching the producer. An Emitter breaks that coupling with the Observer
pattern. Producers broadcast into the Emitter without knowing who is listening,
and consumers subscribe and unsubscribe on their own schedule. The package also
pins down the dispatch contract—ordering, error aggregation, and panic
containment—so callers inherit one consistent behavior instead of re-deriving
it.

# Design

A single Emitter[E] interface describes the contract; two implementations
trade concurrency safety against overhead:

  - Dispatcher[E] is unsynchronized and suited to a single goroutine, such as
    an event loop or actor. It is the lower-overhead option.
  - MutexDispatcher[E] serializes listener registration and removal with a
    mutex, making it safe for concurrent producers and consumers.

Both implementations dispatch to a snapshot of the listeners, in registration
order, so a listener may register or unsubscribe while a dispatch is in
progress and the change takes effect on the next Emit. Failures are contained
per listener: one failing listener does not starve the others, a single error
is returned unchanged, several errors are returned joined, and panics are
recovered and reported as errors rather than propagating.

# Usage

Create an Emitter with NewDispatcher or NewMutexDispatcher, matching the
concurrency model to your callers. Register each listener with On, or OnE when
the listener can fail; the returned Subscription identifies the listener.
Broadcast with Emit, and call Subscription.Off when a listener should stop
receiving events. For one-shot interest, use Once or OnceE, which unsubscribe
automatically after the first delivery. The Example functions below walk
through the full lifecycle.

# Testing

Both Dispatcher and MutexDispatcher are tested against a shared conformance
suite (applyTestEventEmitter) that exercises the full Emitter interface
contract: registration, delivery, unsubscription, once semantics, error
aggregation, and panic containment. Adding a new implementation only requires
wiring it into the suite to inherit the same coverage. The Example functions
serve as executable documentation verified by "go test".
*/
package emitter

import "context"

// Emitter defines the contract implemented by the dispatchers in this package
// for delivering events of type Event to a set of registered listeners.
//
// The On and Once methods accept listeners which cannot fail, while the
// OnE and OnceE variants accept listeners which can report an error. Emit
// delivers the event to every registered listener and returns the errors
// reported by any of them.
type Emitter[Event any] interface {
	// OnE registers l to receive events on subsequent broadcasts, returning a
	// Subscription which can be used to unsubscribe. See Dispatcher.OnE.
	OnE(l ListenerE[Event]) *Subscription[Event]
	// On registers l, which cannot report errors, to receive events on
	// subsequent broadcasts. See Dispatcher.On.
	On(l Listener[Event]) *Subscription[Event]
	// Off removes the given subscription from further broadcasts. It is
	// idempotent.
	Off(s *Subscription[Event])
	// OnceE registers l for a single broadcast; the subscription is removed
	// after the first delivery.
	OnceE(l ListenerE[Event]) *Subscription[Event]
	// Once registers l, which cannot report errors, for a single broadcast;
	// the subscription is removed after the first delivery.
	Once(l Listener[Event]) *Subscription[Event]
	// Emit delivers event to every registered listener in registration order.
	// A listener that returns an error, or panics, does not prevent the
	// remaining listeners from receiving the event; a single failure is
	// returned unchanged and several are returned joined.
	Emit(ctx context.Context, event Event) error
}
