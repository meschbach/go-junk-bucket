/*
Package emitter provides event emission on a 1:n basis.  Typically, these are used for event dispatching patterns.

Emitters come in two flavors:
* Dispatcher designed to work within a single goproc.  It is unsynchronized.
* MutexDispatcher designed to operate in muli-goproc environments made safe via Mutex.

# Lifecycle
A ListenerE is registered via Emitter.On which represents the connection via a Subscription.  When Emitter.Emit is
invoked with an instance of Event will be dispatched.  When a client is no longer interested in events then
Subscription.Off should be invoked to remove the subscription.
*/
package emitter

import "context"

// Emitter defines an interface for event dispatching and listener management with support for context and error handling.
type Emitter[Event any] interface {
	OnE(l ListenerE[Event]) *Subscription[Event]
	On(l Listener[Event]) *Subscription[Event]
	Off(s *Subscription[Event])
	OnceE(l ListenerE[Event]) *Subscription[Event]
	Once(l Listener[Event]) *Subscription[Event]
	Emit(ctx context.Context, event Event) error
}
