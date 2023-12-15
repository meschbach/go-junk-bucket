package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBufferStream(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		t.Parallel()
		assert.Implements(t, (*Sink[int])(nil), NewBuffer[int](3))
		assert.Implements(t, (*Source[float32])(nil), NewBuffer[float32](3))
	})

	t.Run("Given a Buffer with a limit of 3", func(t *testing.T) {
		t.Parallel()
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		bufferSize := 3
		s := NewBuffer[int](bufferSize)
		sinkEvents := AttachSinkVerifier[int](s)

		t.Run("When we add up to the limit", func(t *testing.T) {
			assert.NoError(t, s.Write(ctx, 0))
			assert.NoError(t, s.Write(ctx, 1))
			assert.ErrorIs(t, s.Write(ctx, 2), Full)
			assert.Equal(t, 1, sinkEvents.FullCount, "full buffer triggered event")

			t.Run("Then refuses all additional items", func(t *testing.T) {
				assert.ErrorIs(t, s.Write(ctx, 3), Overflow)
			})

			t.Run("Then it contains all units", func(t *testing.T) {
				assert.Equal(t, []int{0, 1, 2}, s.Output)
			})

			t.Run("And read from", func(t *testing.T) {
				preReadAvailableCount := sinkEvents.AvailableCount
				readOut := make([]int, bufferSize)
				count, err := s.ReadSlice(ctx, readOut)
				assert.NoError(t, err, err)

				t.Run("Then provides all buffered values", func(t *testing.T) {
					assert.Equal(t, []int{0, 1, 2}, readOut, "failed to read expected elements")
					assert.Equal(t, bufferSize, count, "failed to read the expected size")
				})

				assert.Less(t, preReadAvailableCount, sinkEvents.AvailableCount, "expected to dispatch an available event")
			})

			t.Run("And finished", func(t *testing.T) {
				assert.NoError(t, s.Finish(ctx))

				t.Run("Then reading results in End", func(t *testing.T) {
					readOut := make([]int, bufferSize)
					count, err := s.ReadSlice(ctx, readOut)
					assert.Equal(t, 0, count, "expected zero elements, got %#v", readOut[:count])
					assert.ErrorIs(t, err, End)
				})
				t.Run("Then writing results in Finished", func(t *testing.T) {
					e := s.Write(ctx, 5)
					assert.ErrorIs(t, e, Done)
				})
			})
		})
	})
}
