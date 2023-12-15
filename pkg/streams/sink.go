package streams

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
)

// Sink represents a way to write elements to an abstract destination.  Could be a memory buffer, through a channel,
// a database, etc.  As a key counterpart to Source, a Sink provides back pressure semantics to allow buffers to be
// primed.
type Sink[T any] interface {
	//Write offers a given value v to the Sink.  Standard errors are as follows:
	// * Done - If the stream has finished.  No further writes will be consumed
	// * Full - Written element has filled the output buffers and further writes will result in Overflow
	// * Overflow - Element has been rejected as the sink has too many pending elements to be written
	// * Done - Element has been rejected as the sink is draining
	Write(ctx context.Context, v T) error

	// Finish instructs the sink to no longer accept writes, wait for internal buffers to drain, and finish.
	// Source must emit at least the following events:
	// * Finishing - Indicates the sink is no longer accepting new writes
	// * Finished - Indicates the sink has completed flushing the internal buffer
	Finish(ctx context.Context) error

	// SinkEvents provides access to the possible events to be emitted by the sink
	SinkEvents() *SinkEvents[T]

	// Resume instructs the Sink to resume draining internal buffers and accept writes again.
	// Standard errors:
	// * Done - The stream is finishing or finishing and is no longer accepting writes.
	Resume(ctx context.Context) error
}

// SinkEvents are the available events to interact with the sink
type SinkEvents[T any] struct {
	// Writable is dispatched to indicate space is available in a Sink.  This is typically done at the end of any
	// logical grouping or batch which would the sink is drained to.  This allows more efficient pushing.
	//
	// todo: consider merging with drained based on how clients will react.
	Available emitter.Dispatcher[SinkAvailableEvent[T]]

	// Drained is dispatched when the sink is depleted with no additional elements available.
	// todo: consider merging with available based on how clients will react.
	Drained emitter.Dispatcher[Sink[T]]

	// Full is dispatched when the sink is no longer accepting elements.
	Full emitter.Dispatcher[Sink[T]]

	// Finishing is emitted when the stream is waiting for all elements to be flushed but not yet finished.  No further
	// elements will be accepted by the stream
	Finishing emitter.Dispatcher[Sink[T]]

	// Finished indicates the stream has completed flushing its internal buffer and has no further work as a sink to
	// complete.
	Finished emitter.Dispatcher[Sink[T]]
}

type SinkAvailableEvent[T any] struct {
	// Space indicates how many elements can be written before Overflow is provided.  This is advisory as other handlers
	// may have already written data.  Sink.Write is the authoritative source of data for this
	Space int
	// Sink is the originating entity for this event.
	Sink Sink[T]
}

// Overflow indicates the stream would overflow with the given item.
var Overflow = errors.New("stream overflow")

// Full represents a writable stream whose internal buffers which have been filled.  A Full error provides a clear
// signal for backpressure on the emitting source.
var Full = errors.New("stream full")
