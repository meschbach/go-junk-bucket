package fx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	t.Run("Given a nil slice", func(t *testing.T) {
		var input []int

		t.Run("When split", func(t *testing.T) {
			left, right := Split(input, func(e int) bool {
				return true
			})

			t.Run("Then both slices are empty", func(t *testing.T) {
				assert.Len(t, left, 0)
				assert.Len(t, right, 0)
			})
		})
	})

	t.Run("Given an empty slice", func(t *testing.T) {
		input := []int{}

		t.Run("When split", func(t *testing.T) {
			left, right := Split(input, func(e int) bool {
				return true
			})

			t.Run("Then both slices are empty", func(t *testing.T) {
				assert.Len(t, left, 0)
				assert.Len(t, right, 0)
			})
		})
	})

	t.Run("Given a single element which matches", func(t *testing.T) {
		input := []int{42}

		t.Run("When split", func(t *testing.T) {
			left, right := Split(input, func(e int) bool {
				return e == 42
			})

			t.Run("Then the element is in the left slice", func(t *testing.T) {
				assert.Equal(t, []int{42}, left)
			})
			t.Run("And the right slice is empty", func(t *testing.T) {
				assert.Len(t, right, 0)
			})
		})
	})

	t.Run("Given a single element which does not match", func(t *testing.T) {
		input := []int{42}

		t.Run("When split", func(t *testing.T) {
			left, right := Split(input, func(e int) bool {
				return e == 0
			})

			t.Run("Then the left slice is empty", func(t *testing.T) {
				assert.Len(t, left, 0)
			})
			t.Run("And the element is in the right slice", func(t *testing.T) {
				assert.Equal(t, []int{42}, right)
			})
		})
	})

	t.Run("Given elements that all match", func(t *testing.T) {
		input := []int{2, 4, 6}

		t.Run("When split", func(t *testing.T) {
			left, right := Split(input, func(e int) bool {
				return e%2 == 0
			})

			t.Run("Then all elements are in the left slice", func(t *testing.T) {
				assert.Equal(t, []int{2, 4, 6}, left)
			})
			t.Run("And the right slice is empty", func(t *testing.T) {
				assert.Len(t, right, 0)
			})
		})
	})

	t.Run("Given elements that none match", func(t *testing.T) {
		input := []int{1, 3, 5}

		t.Run("When split", func(t *testing.T) {
			left, right := Split(input, func(e int) bool {
				return e%2 == 0
			})

			t.Run("Then the left slice is empty", func(t *testing.T) {
				assert.Len(t, left, 0)
			})
			t.Run("And all elements are in the right slice", func(t *testing.T) {
				assert.Equal(t, []int{1, 3, 5}, right)
			})
		})
	})

	t.Run("Given a set of numbers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		t.Run("When split less than 5", func(t *testing.T) {
			lt, gt := Split[int](input, func(e int) bool {
				return e < 5
			})

			assert.Equal(t, []int{1, 2, 3, 4}, lt)
			assert.Equal(t, []int{5, 6, 7, 8, 9, 10}, gt)
		})

		t.Run("When split", func(t *testing.T) {
			original := []int{1, 2, 3, 4, 5}
			_, _ = Split(original, func(e int) bool {
				return e > 2
			})

			t.Run("Then the original slice is unmodified", func(t *testing.T) {
				assert.Equal(t, []int{1, 2, 3, 4, 5}, original)
			})
		})

		t.Run("When split into evens and odds", func(t *testing.T) {
			evens, odds := Split(input, func(e int) bool {
				return e%2 == 0
			})

			t.Run("Then both outputs preserve the original order", func(t *testing.T) {
				assert.Equal(t, []int{2, 4, 6, 8, 10}, evens)
				assert.Equal(t, []int{1, 3, 5, 7, 9}, odds)
			})
		})
	})
}
