package fx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("Given an empty array", func(t *testing.T) {
		var input []int64

		t.Run("When mapped", func(t *testing.T) {
			output := Map(input, Identity[int64])

			t.Run("Then it results in an empty slice", func(t *testing.T) {
				assert.Len(t, output, 0)
			})
		})
	})

	t.Run("Given a set of input values", func(t *testing.T) {
		input := []int{1, 2, 4, 8, 16}

		t.Run("When mapped with to string", func(t *testing.T) {
			output := Map(input, func(i int) string {
				return fmt.Sprintf("%d", i)
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
	})
}
