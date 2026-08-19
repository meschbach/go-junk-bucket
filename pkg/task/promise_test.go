package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromise(t *testing.T) {
	t.Run("Given a new promise with a registered success handler", func(t *testing.T) {
		p := Promise[int]{}
		lastInvokedValue := -1
		p.Then(t.Context(), func(ctx context.Context, event int) {
			lastInvokedValue = event
		})

		t.Run("When the promise is completed", func(t *testing.T) {
			p.Success(t.Context(), 42)

			t.Run("Then the handler is invoked", func(t *testing.T) {
				assert.Equal(t, 42, lastInvokedValue)
			})
		})

		t.Run("When attempting to complete again", func(t *testing.T) {
			t.Run("Then it panics", func(t *testing.T) {
				assert.Panics(t, func() {
					p.Success(t.Context(), 46)
				}, "second success should panic")
			})
		})
	})
}
