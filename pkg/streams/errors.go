package streams

import "errors"

// Done represents a writable stream which has been closed.  A Sink may return this in the time between being instructed
// to close and still draining all elements from their buffer
var Done = errors.New("stream done")

// End indicates a source has read all elements available on the stream.
var End = errors.New("stream end")
