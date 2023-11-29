package streams

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/emitter"
)

// SourceEvents are the set of events which may be emitted from a source to signal various conditions and states.
type SourceEvents[T any] struct {
	Data emitter.Dispatcher[T]
	End  emitter.Dispatcher[Source[T]]
}

type Source[T any] interface {
	SourceEvents() *SourceEvents[T]
	ReadSlice(ctx context.Context, to []T) (int, error)
	// Resume begins emitting events of T until pushback is received from an event.
	Resume(ctx context.Context) error
	Pause(ctx context.Context) error
}
