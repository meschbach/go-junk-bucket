package reactive

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFixedSlice(t *testing.T) {
	t.Run("Given a fixed slice", func(t *testing.T) {
		f := FromSlice([]int{0, 1, 1, 2, 3, 4, 5})

		t.Run("When reading with a smaller buffer", func(t *testing.T) {
			target := make([]int, 4)
			count, err := f.ReadSlice(context.Background(), target)
			require.NoError(t, err)
			assert.Equal(t, 4, count)

			t.Run("Then it provides all values", func(t *testing.T) {
				assert.Equal(t, []int{0, 1, 1, 2}, target)
			})

			t.Run("And the buffer is read beyond the end", func(t *testing.T) {
				target := make([]int, 4)
				count, err := f.ReadSlice(context.Background(), target)
				t.Run("Then is returns the short read state", func(t *testing.T) {
					assert.NoError(t, err)
					assert.Equal(t, 3, count)

					//note: the zero is added by golang through its automagic
					assert.Equal(t, []int{3, 4, 5, 0}, target)
				})
			})
		})
	})
}
