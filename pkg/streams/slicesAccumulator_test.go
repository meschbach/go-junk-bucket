package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSliceAccumulator(t *testing.T) {
	t.Run("Given a new slice accumulator", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		s := NewSliceAccumulator[int]()
		hasFinished := false
		s.Events.OnFinished.On(func(ctx context.Context, event Sink[int]) {
			hasFinished = true
		})

		t.Run("When written too", func(t *testing.T) {
			assert.NoError(t, s.Write(ctx, 0))
			assert.NoError(t, s.Write(ctx, 1))
			assert.NoError(t, s.Write(ctx, 2))

			t.Run("Then all values have been stored", func(t *testing.T) {
				assert.Equal(t, []int{0, 1, 2}, s.Output)
			})

			t.Run("Then has not finished", func(t *testing.T) {
				assert.False(t, hasFinished, "has finished")
			})
		})

		t.Run("When the stream is finished", func(t *testing.T) {
			assert.NoError(t, s.Finish(ctx))

			t.Run("Then an event is dispatched", func(t *testing.T) {
				assert.True(t, hasFinished, "has finished")
			})

			t.Run("Then the stream does not accept further writes", func(t *testing.T) {
				assert.ErrorIs(t, s.Write(ctx, 3), Done)

				t.Run("And the stream does not include the value", func(t *testing.T) {
					assert.Equal(t, []int{0, 1, 2}, s.Output)
				})
			})
		})
	})

	t.Run("SliceAccumulator is Sink", func(t *testing.T) {
		s := NewSliceAccumulator[int]()
		assert.Implements(t, (*Sink[int])(nil), s)
	})
}
