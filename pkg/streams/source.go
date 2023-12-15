package streams

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
)

var UnderRun = errors.New("stream under run")

// SourceEvents are the set of events which may be emitted from a source to signal various conditions and states.
type SourceEvents[T any] struct {
	Data emitter.Dispatcher[T]
	End  emitter.Dispatcher[Source[T]]
}

type Source[T any] interface {
	SourceEvents() *SourceEvents[T]
	// ReadSlice reads up to len(to) elements from the source, returning the count of elements read into the buffer and
	// any problems encountered while reading.  In the case 0 elements are available the source should return
	// UnderRun as the error.
	ReadSlice(ctx context.Context, to []T) (count int, problem error)
	// Resume begins emitting events of T until pushback is received from an event.
	Resume(ctx context.Context) error
	Pause(ctx context.Context) error
}
