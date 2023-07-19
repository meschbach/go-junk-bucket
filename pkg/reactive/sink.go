package reactive

import "context"

type SinkEvents[T any] struct {
	OnDrain    EventEmitter[Sink[T]]
	OnFinished EventEmitter[Sink[T]]
}

type Sink[T any] interface {
	//Write offers a given value v to the Sink.  Standard errors are as follows:
	// * Done - If the stream has finished.  No further writes will be consumed
	// * Full - Buffers are currently full and unable to accept another value.  Listen for OnDrain for additional capacity.
	Write(ctx context.Context, v T) error
	Finish(ctx context.Context) error
	SinkEvents() *SinkEvents[T]

	Resume(ctx context.Context) error
}
