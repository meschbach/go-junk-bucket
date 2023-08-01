package emitter

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEventEmitter(t *testing.T) {
	t.Run("Given an event emitter", func(t *testing.T) {
		e := Dispatcher[int]{}
		t.Run("When a listener is registered", func(t *testing.T) {
			received := -1
			subscription := e.On(func(ctx context.Context, event int) {
				received = event
			})

			t.Run("And an event is dispatched", func(t *testing.T) {
				assert.NoError(t, e.Emit(context.Background(), 4))
				t.Run("Then it receives future events", func(t *testing.T) {
					assert.Equal(t, 4, received)
				})
			})

			t.Run("And the listener is unsubscribed", func(t *testing.T) {
				e.Off(subscription)

				t.Run("Then no further events are dispatched", func(t *testing.T) {
					assert.NoError(t, e.Emit(context.Background(), 5))
					assert.Equal(t, 5, received)
				})
			})
		})
	})

	t.Run("Given a dispatcher registered with a handler which adds another", func(t *testing.T) {
		e := Dispatcher[int]{}
		lastOuterValue := -1
		immediatelyCalled := -1
		e.On(func(ctx context.Context, event int) {
			lastOuterValue = event
			e.Once(func(ctx context.Context, event int) {
				immediatelyCalled = event
			})
		})

		t.Run("When initially dispatching", func(t *testing.T) {
			require.NoError(t, e.Emit(context.Background(), 42))
			t.Run("Then the new handler is not called", func(t *testing.T) {
				assert.Equal(t, 42, lastOuterValue)
				assert.Equal(t, -1, immediatelyCalled)
			})
		})

		t.Run("When dispatched a second time", func(t *testing.T) {
			require.NoError(t, e.Emit(context.Background(), 46))
			t.Run("Then both handlers receive the value", func(t *testing.T) {
				assert.Equal(t, 46, lastOuterValue)
				assert.Equal(t, 46, immediatelyCalled)
			})
		})
	})

	t.Run("Given an event listener which generates an error", func(t *testing.T) {
		e := Dispatcher[int]{}
		todo := errors.New("todo")
		secondListener := -1

		e.OnE(func(ctx context.Context, event int) error {
			return todo
		})
		e.On(func(ctx context.Context, event int) {
			secondListener = event
		})

		t.Run("When an event is emitted", func(t *testing.T) {
			exampleValue := 128
			problem := e.Emit(context.Background(), exampleValue)

			t.Run("Then the error is returned", func(t *testing.T) {
				assert.ErrorIs(t, problem, todo)
			})
			t.Run("Then the second event handler is still called", func(t *testing.T) {
				assert.Equal(t, exampleValue, secondListener)
			})
		})
	})
}
