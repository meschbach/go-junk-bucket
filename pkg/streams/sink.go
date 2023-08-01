package streams

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
)

type SinkEvents[T any] struct {
	OnDrain    emitter.Dispatcher[Sink[T]]
	OnFinished emitter.Dispatcher[Sink[T]]
}

type Sink[T any] interface {
	//Write offers a given value v to the Sink.  Standard errors are as follows:
	// * Done - If the stream has finished.  No further writes will be consumed
	// * Full - Buffers are currently full and unable to accept another value.  Listen for OnDrain for additional capacity.
	Write(ctx context.Context, v T) error
	// Finish instructs the sink to no longer accept writes and flush internal buffers and clean up.
	Finish(ctx context.Context) error
	SinkEvents() *SinkEvents[T]

	// Resume instructs the Sink to resume draining internal buffers and accept writes again.
	Resume(ctx context.Context) error
}
