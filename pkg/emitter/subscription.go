package emitter

type offable[E any] interface {
	Off(s *Subscription[E])
}

// Subscription represents a single listener bound to an Emitter. It is
// returned by the registration methods (On, OnE, Once, OnceE) and used to
// unsubscribe the listener via Off.
type Subscription[E any] struct {
	from   offable[E]
	target ListenerE[E]
}

// Off unregisters the current subscription from receiving further events from
// the Emitter it belongs to. It is idempotent: calling Off again after the
// subscription has already been removed has no effect.
func (s *Subscription[E]) Off() {
	s.from.Off(s)
}
