package reactors

import (
	"context"
	"errors"
)

// RunChannelActor will run a new reactor with the given state until the given context is complete.
func RunChannelActor[E any](ctx context.Context, state E) *Channel[E] {
	reactor, queue := NewChannel[E](32)
	go func() {
		for {
			select {
			case e := <-queue:
				if err := reactor.Tick(ctx, e, state); err != nil { //todo: better error feedback mechanism?
					if errors.Is(err, context.Canceled) {
						//do nothing
					} else {
						panic(err)
					}
				}
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					if errors.Is(err, context.Canceled) { //todo: better error feedback mechanism?
						//do nothing
					} else {
						panic(err)
					}
				}
			}
		}
	}()
	return reactor
}
