package reactors

import (
	"context"
	"errors"
)

func RunChannelActor[E any](ctx context.Context, s E) *Channel[E] {
	reactor, _ := NewChannel[E](32)
	go func() {
		_, err := reactor.ConsumeAll(ctx, s)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				//do nothing
			} else {
				panic(err)
			}
		}
	}()
	return reactor
}
