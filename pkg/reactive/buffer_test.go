package reactive

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLimitedSliceAccumulator(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		assert.Implements(t, (*Sink[int])(nil), NewBuffer[int](3))
		assert.Implements(t, (*Source[float32])(nil), NewBuffer[float32](3))
	})

	t.Run("Given a Limited accumulator of 3", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		bufferSize := 3
		s := NewBuffer[int](bufferSize)
		t.Run("When we add up to the limit", func(t *testing.T) {
			assert.NoError(t, s.Write(ctx, 0))
			assert.NoError(t, s.Write(ctx, 1))
			assert.NoError(t, s.Write(ctx, 2))

			t.Run("Then it contains all units", func(t *testing.T) {
				assert.Equal(t, []int{0, 1, 2}, s.Output)
			})

			t.Run("Then refuses all additional items", func(t *testing.T) {
				assert.ErrorIs(t, s.Write(ctx, 3), Full)
			})

			t.Run("And read from", func(t *testing.T) {
				readOut := make([]int, bufferSize)
				count, err := s.ReadSlice(ctx, readOut)
				assert.NoError(t, err)

				t.Run("Then all elements read", func(t *testing.T) {
					assert.Equal(t, bufferSize, count)
				})

				t.Run("Then provides all buffered values", func(t *testing.T) {
					assert.Equal(t, []int{0, 1, 2}, readOut)
				})
			})
			t.Run("And finished", func(t *testing.T) {
				assert.NoError(t, s.Finish(ctx))

				t.Run("Then reading results in End", func(t *testing.T) {
					readOut := make([]int, bufferSize)
					count, err := s.ReadSlice(ctx, readOut)
					assert.Equal(t, 0, count)
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
