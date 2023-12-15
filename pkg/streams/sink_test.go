package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SinkVerificationHarness[T any] struct {
	AvailableCount int
	DrainedCount   int
	FinishingCount int
	FinishedCount  int
	FullCount      int
}

func (s SinkVerificationHarness[T]) AssertOpen(t *testing.T) {
	assert.Equal(t, 0, s.FinishingCount)
	assert.Equal(t, 0, s.FinishedCount)
}

func AttachSinkVerifier[T any](to Sink[T]) *SinkVerificationHarness[T] {
	harness := &SinkVerificationHarness[T]{}
	emitter := to.SinkEvents()
	emitter.Drained.On(func(ctx context.Context, event Sink[T]) {
		harness.DrainedCount++
	})
	emitter.Finished.On(func(ctx context.Context, event Sink[T]) {
		harness.FinishedCount++
	})
	emitter.Finishing.On(func(ctx context.Context, event Sink[T]) {
		harness.FinishingCount++
	})
	emitter.Full.On(func(ctx context.Context, event Sink[T]) {
		harness.FullCount++
	})
	emitter.Available.On(func(ctx context.Context, event SinkAvailableEvent[T]) {
		harness.AvailableCount++
	})
	return harness
}
