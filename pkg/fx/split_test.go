package fx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplit(t *testing.T) {
	t.Run("Given a set of numbers", func(t *testing.T) {
		t.Run("When split less than 5", func(t *testing.T) {
			example := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

			lt, gt := Split[int](example, func(e int) bool {
				return e < 5
			})

			assert.Equal(t, []int{1, 2, 3, 4}, lt)
			assert.Equal(t, []int{5, 6, 7, 8, 9, 10}, gt)
		})
	})
}
