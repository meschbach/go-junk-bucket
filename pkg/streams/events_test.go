package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventEmitter(t *testing.T) {
	t.Run("Given an event emitter", func(t *testing.T) {
		e := EventEmitter[int]{}
		t.Run("When a listener is registered", func(t *testing.T) {
			received := -1
			e.On(func(ctx context.Context, event int) error {
				received = event
				return nil
			})

			t.Run("And an event is dispatched", func(t *testing.T) {
				assert.NoError(t, e.Emit(context.Background(), 4))
				t.Run("Then it receives future events", func(t *testing.T) {
					assert.Equal(t, 4, received)
				})
			})
		})
	})
}
