package reactive

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLimitedSliceAccumulator(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		assert.Implements(t, (*Sink[int])(nil), NewBuffer[int](3))
	})

	t.Run("Given a Limited accumulator of 3", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		s := NewBuffer[int](3)
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
		})
	})
}
