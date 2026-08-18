package fx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {
	t.Run("Given an integer", func(t *testing.T) {
		t.Run("When passed through Identity", func(t *testing.T) {
			result := Identity(42)

			t.Run("Then the value is returned unchanged", func(t *testing.T) {
				assert.Equal(t, 42, result)
			})
		})
	})

	t.Run("Given a string", func(t *testing.T) {
		t.Run("When passed through Identity", func(t *testing.T) {
			result := Identity("hello")

			t.Run("Then the value is returned unchanged", func(t *testing.T) {
				assert.Equal(t, "hello", result)
			})
		})
	})

	t.Run("Given a nil pointer", func(t *testing.T) {
		t.Run("When passed through Identity", func(t *testing.T) {
			var p *int
			result := Identity(p)

			t.Run("Then nil is returned", func(t *testing.T) {
				assert.Nil(t, result)
			})
		})
	})

	t.Run("Given a slice", func(t *testing.T) {
		t.Run("When passed through Identity", func(t *testing.T) {
			input := []int{1, 2, 3}
			result := Identity(input)

			t.Run("Then the same slice is returned", func(t *testing.T) {
				assert.Equal(t, input, result)
			})
		})
	})
}
