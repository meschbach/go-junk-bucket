package emitter

import "context"

// ListenerE is the typing for the function to receive events, possibly resulting in error on result
type ListenerE[E any] func(ctx context.Context, event E) error

// Listener is the typing for the function to receive events not resulting in an error
type Listener[E any] func(ctx context.Context, event E)
