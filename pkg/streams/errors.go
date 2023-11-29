package streams

import "errors"

// Done represents a writable stream which has been closed.  A Sink may return this in the time between being instructed
// to close and still draining all elements from their buffer
var Done = errors.New("stream done")

// Full represents a writable stream whose internal buffers which have been filled.  A Full error provides a clear
// signal for backpressure on the emitting source.
var Full = errors.New("stream full")

// End indicates a source has read all elements available on the stream.
var End = errors.New("stream end")
