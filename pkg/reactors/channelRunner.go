package reactors

import (
	"context"
	"errors"
)

// RunChannelActor starts a goroutine running a [Channel] reactor with the given
// initial state. The reactor processes events until ctx is canceled.
//
// Errors other than context.Canceled cause a panic. The returned [Channel] can
// be used to schedule work into the running reactor.
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
