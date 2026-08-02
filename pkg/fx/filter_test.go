package fx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Run("Given an empty slice", func(t *testing.T) {
		var input []int64

		t.Run("When filtered", func(t *testing.T) {
			output := Filter[int64](input, func(e int64) bool {
				return true
			})

			t.Run("Then it results in an empty slice", func(t *testing.T) {
				assert.Len(t, output, 0)
			})
		})
	})

	t.Run("Given a set of numbers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		t.Run("When filtered for evens", func(t *testing.T) {
			output := Filter[int](input, func(e int) bool {
				return e%2 == 0
			})

			t.Run("Then only the even values are kept", func(t *testing.T) {
				assert.Equal(t, []int{2, 4, 6, 8, 10}, output)
			})
		})

		t.Run("When filtered for values less than 5", func(t *testing.T) {
			output := Filter[int](input, func(e int) bool {
				return e < 5
			})

			t.Run("Then only the values less than 5 are kept", func(t *testing.T) {
				assert.Equal(t, []int{1, 2, 3, 4}, output)
			})
		})

		t.Run("When filtered with a test that is always false", func(t *testing.T) {
			output := Filter[int](input, func(e int) bool {
				return false
			})

			t.Run("Then it results in an empty slice", func(t *testing.T) {
				assert.Len(t, output, 0)
			})
		})

		t.Run("When filtered with a test that is always true", func(t *testing.T) {
			output := Filter[int](input, func(e int) bool {
				return true
			})

			t.Run("Then all values are kept in the original order", func(t *testing.T) {
				assert.Equal(t, input, output)
			})
		})
	})
}
