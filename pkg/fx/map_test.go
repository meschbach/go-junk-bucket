package fx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("Given a nil slice", func(t *testing.T) {
		var input []int

		t.Run("When mapped", func(t *testing.T) {
			output := Map(input, func(i int) string {
				return fmt.Sprintf("%d", i)
			})

			t.Run("Then it results in an empty slice", func(t *testing.T) {
				assert.Len(t, output, 0)
			})
		})
	})

	t.Run("Given an empty slice", func(t *testing.T) {
		var input []int64

		t.Run("When mapped", func(t *testing.T) {
			output := Map(input, Identity[int64])

			t.Run("Then it results in an empty slice", func(t *testing.T) {
				assert.Len(t, output, 0)
			})
		})
	})

	t.Run("Given a single element", func(t *testing.T) {
		input := []int{5}

		t.Run("When mapped to a string", func(t *testing.T) {
			output := Map(input, func(i int) string {
				return fmt.Sprintf("%d", i)
			})

			t.Run("Then it results in a single-element slice", func(t *testing.T) {
				assert.Equal(t, []string{"5"}, output)
			})
		})
	})

	t.Run("Given a set of input values", func(t *testing.T) {
		input := []int{1, 2, 4, 8, 16}

		t.Run("When mapped to strings", func(t *testing.T) {
			output := Map(input, func(i int) string {
				return fmt.Sprintf("%d", i)
			})

			t.Run("Then the output has the same length as the input", func(t *testing.T) {
				assert.Len(t, output, len(input))
			})

			t.Run("Then it results in the expected values", func(t *testing.T) {
				if assert.Len(t, output, 5) {
					assert.Equal(t, "1", output[0])
					assert.Equal(t, "2", output[1])
					assert.Equal(t, "4", output[2])
					assert.Equal(t, "8", output[3])
					assert.Equal(t, "16", output[4])
				}
			})
		})

		t.Run("When mapped with a same-type transform", func(t *testing.T) {
			output := Map(input, func(i int) int {
				return i * 2
			})

			t.Run("Then it results in the doubled values", func(t *testing.T) {
				assert.Equal(t, []int{2, 4, 8, 16, 32}, output)
			})
		})

		t.Run("When mapped", func(t *testing.T) {
			original := []int{1, 2, 3}
			_ = Map(original, func(i int) string {
				return fmt.Sprintf("%d", i)
			})

			t.Run("Then the original slice is unmodified", func(t *testing.T) {
				assert.Equal(t, []int{1, 2, 3}, original)
			})
		})
	})
}
