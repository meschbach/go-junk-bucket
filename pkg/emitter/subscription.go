package emitter

type offable[E any] interface {
	Off(s *Subscription[E])
}

// Subscription represents a single listener bound to hear events
type Subscription[E any] struct {
	from   offable[E]
	target ListenerE[E]
}

// Off unregisters the current subscription from receiving further events from the dispatcher it belongs to.
func (s *Subscription[E]) Off() {
	s.from.Off(s)
}
