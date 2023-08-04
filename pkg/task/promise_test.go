package task

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPromise(t *testing.T) {
	t.Run("Given a new promise with a registered success handler", func(t *testing.T) {
		p := Promise[int]{}
		lastInvokedValue := -1
		p.Then(context.Background(), func(ctx context.Context, event int) {
			lastInvokedValue = event
		})

		t.Run("When the promise is completed", func(t *testing.T) {
			p.Success(context.Background(), 42)

			t.Run("Then the handler is invoked", func(t *testing.T) {
				assert.Equal(t, 42, lastInvokedValue)
			})
		})

		t.Run("When attempting to complete again", func(t *testing.T) {
			t.Run("Then it panics", func(t *testing.T) {
				assert.Panics(t, func() {
					p.Success(context.Background(), 46)
				}, "second success should panic")
			})
		})
	})
}
