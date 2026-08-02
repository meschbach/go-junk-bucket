package emitter

import "context"

// ListenerE is a function which receives an event and can report an error
// preventing successful processing. Register via Emitter.OnE or Emitter.OnceE.
type ListenerE[E any] func(ctx context.Context, event E) error

// Listener is a function which receives an event and cannot report an error.
// It is the convenient form of ListenerE for listeners which always succeed.
// Register via Emitter.On or Emitter.Once.
type Listener[E any] func(ctx context.Context, event E)
