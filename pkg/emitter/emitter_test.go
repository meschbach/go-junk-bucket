package emitter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EmitterConstructor func() Emitter[int]

// applyTestEventEmitter runs the full conformance suite against the given
// emitter implementation, covering the documented lifecycle as well as the
// unsubscribe, once, error, and panic guarantees.
func applyTestEventEmitter(t *testing.T, newEmitter EmitterConstructor) {
	applyTestEventEmitterLifecycle(t, newEmitter)
	applyTestEventEmitterErrors(t, newEmitter)
	applyTestEventEmitterGuarantees(t, newEmitter)
	applyTestEventEmitterEdgeCases(t, newEmitter)
	applyTestEventEmitterDispatchMutations(t, newEmitter)
}

func applyTestEventEmitterLifecycle(t *testing.T, newEmitter EmitterConstructor) {
	t.Run("Given an event emitter", func(t *testing.T) {
		e := newEmitter()
		t.Run("When a listener is registered", func(t *testing.T) {
			received := -1
			subscription := e.On(func(_ context.Context, event int) {
				received = event
			})

			t.Run("And an event is dispatched", func(t *testing.T) {
				require.NoError(t, e.Emit(t.Context(), 4))
				t.Run("Then it receives future events", func(t *testing.T) {
					assert.Equal(t, 4, received)
				})
			})

			t.Run("And the listener is unsubscribed", func(t *testing.T) {
				e.Off(subscription)

				t.Run("Then no further events are dispatched", func(t *testing.T) {
					require.NoError(t, e.Emit(t.Context(), 5))
					assert.Equal(t, 4, received)
				})
			})
		})
	})

	t.Run("Given a dispatcher registered with a handler which adds another", func(t *testing.T) {
		e := newEmitter()
		lastOuterValue := -1
		immediatelyCalled := -1
		e.On(func(_ context.Context, event int) {
			lastOuterValue = event
			e.Once(func(_ context.Context, event int) {
				immediatelyCalled = event
			})
		})

		t.Run("When initially dispatching", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 42))
			t.Run("Then the new handler is not called", func(t *testing.T) {
				assert.Equal(t, 42, lastOuterValue)
				assert.Equal(t, -1, immediatelyCalled)
			})
		})

		t.Run("When dispatched a second time", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 46))
			t.Run("Then both handlers receive the value", func(t *testing.T) {
				assert.Equal(t, 46, lastOuterValue)
				assert.Equal(t, 46, immediatelyCalled)
			})
		})
	})
}

func applyTestEventEmitterErrors(t *testing.T, newEmitter EmitterConstructor) {
	t.Run("Given an event listener which generates an error", func(t *testing.T) {
		e := newEmitter()
		todo := errors.New("todo")
		secondListener := -1

		e.OnE(func(_ context.Context, _ int) error {
			return todo
		})
		e.On(func(_ context.Context, event int) {
			secondListener = event
		})

		t.Run("When an event is emitted", func(t *testing.T) {
			exampleValue := 128
			problem := e.Emit(t.Context(), exampleValue)

			t.Run("Then the error is returned", func(t *testing.T) {
				assert.ErrorIs(t, problem, todo)
			})
			t.Run("Then the second event handler is still called", func(t *testing.T) {
				assert.Equal(t, exampleValue, secondListener)
			})
		})
	})

	t.Run("Given a single listener which errors", func(t *testing.T) {
		e := newEmitter()
		todo := errors.New("todo")
		e.OnE(func(_ context.Context, _ int) error {
			return todo
		})

		t.Run("When an event is emitted", func(t *testing.T) {
			err := e.Emit(t.Context(), 1)

			t.Run("Then the original error is returned unchanged", func(t *testing.T) {
				require.ErrorIs(t, err, todo)
				assert.Same(t, todo, err)
			})
		})
	})
}

func applyTestEventEmitterGuarantees(t *testing.T, newEmitter EmitterConstructor) {
	t.Run("Given two listeners", func(t *testing.T) {
		e := newEmitter()
		first := -1
		second := -1
		firstSub := e.On(func(_ context.Context, event int) {
			first = event
		})
		e.On(func(_ context.Context, event int) {
			second = event
		})

		t.Run("When the first is unsubscribed", func(t *testing.T) {
			firstSub.Off()
			require.NoError(t, e.Emit(t.Context(), 7))

			t.Run("Then it no longer receives events", func(t *testing.T) {
				assert.Equal(t, -1, first)
			})
			t.Run("And the other listener still does", func(t *testing.T) {
				assert.Equal(t, 7, second)
			})
		})
	})

	t.Run("Given a once listener", func(t *testing.T) {
		e := newEmitter()
		calls := 0
		e.Once(func(_ context.Context, _ int) {
			calls++
		})

		t.Run("When events are emitted repeatedly", func(t *testing.T) {
			for i := 0; i < 3; i++ {
				require.NoError(t, e.Emit(t.Context(), i))
			}

			t.Run("Then it is called exactly once", func(t *testing.T) {
				assert.Equal(t, 1, calls)
			})
		})
	})

	t.Run("Given a once listener which panics", func(t *testing.T) {
		e := newEmitter()
		calls := 0
		e.Once(func(_ context.Context, _ int) {
			calls++
			panic("boom")
		})

		t.Run("When an event is emitted twice", func(t *testing.T) {
			require.Error(t, e.Emit(t.Context(), 1))
			require.NoError(t, e.Emit(t.Context(), 2))

			t.Run("Then the once listener fires only once", func(t *testing.T) {
				assert.Equal(t, 1, calls)
			})
		})
	})

	t.Run("Given a listener which panics", func(t *testing.T) {
		e := newEmitter()
		panicErr := errors.New("boom")
		after := -1
		e.OnE(func(_ context.Context, _ int) error {
			panic(panicErr)
		})
		e.On(func(_ context.Context, event int) {
			after = event
		})

		t.Run("When an event is emitted", func(t *testing.T) {
			problem := e.Emit(t.Context(), 9)

			t.Run("Then the panic is reported as an error", func(t *testing.T) {
				require.ErrorIs(t, problem, panicErr)
				assert.Same(t, panicErr, problem)
			})
			t.Run("And the later listener still receives the event", func(t *testing.T) {
				assert.Equal(t, 9, after)
			})
		})
	})
}

func applyTestEventEmitterEdgeCases(t *testing.T, newEmitter EmitterConstructor) {
	t.Run("Given an emitter with no listeners", func(t *testing.T) {
		e := newEmitter()

		t.Run("When an event is emitted", func(t *testing.T) {
			err := e.Emit(t.Context(), 42)

			t.Run("Then no error is returned", func(t *testing.T) {
				assert.NoError(t, err)
			})
		})
	})

	t.Run("Given two listeners which both error", func(t *testing.T) {
		e := newEmitter()
		firstErr := errors.New("first")
		secondErr := errors.New("second")

		e.OnE(func(_ context.Context, _ int) error { return firstErr })
		e.OnE(func(_ context.Context, _ int) error { return secondErr })

		t.Run("When an event is emitted", func(t *testing.T) {
			err := e.Emit(t.Context(), 1)

			t.Run("Then both errors are present in the result", func(t *testing.T) {
				require.Error(t, err)
				assert.ErrorIs(t, err, firstErr)
				assert.ErrorIs(t, err, secondErr)
			})
		})
	})

	t.Run("Given a listener which panics with a non-error value", func(t *testing.T) {
		e := newEmitter()
		e.OnE(func(_ context.Context, _ int) error {
			panic("string panic")
		})

		t.Run("When an event is emitted", func(t *testing.T) {
			err := e.Emit(t.Context(), 1)

			t.Run("Then the panic is wrapped in an error", func(t *testing.T) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "string panic")
				assert.Contains(t, err.Error(), "listener panic")
			})
		})
	})

	t.Run("Given a subscription", func(t *testing.T) {
		e := newEmitter()
		calls := 0
		sub := e.On(func(_ context.Context, _ int) {
			calls++
		})

		t.Run("When Off is called twice", func(t *testing.T) {
			sub.Off()
			sub.Off()

			t.Run("Then no panic occurs and the listener is removed", func(t *testing.T) {
				require.NoError(t, e.Emit(t.Context(), 1))
				assert.Equal(t, 0, calls)
			})
		})
	})

	t.Run("Given a OnceE listener which errors", func(t *testing.T) {
		e := newEmitter()
		onceErr := errors.New("once error")
		calls := 0

		e.OnceE(func(_ context.Context, _ int) error {
			calls++
			return onceErr
		})

		t.Run("When an event is emitted twice", func(t *testing.T) {
			err := e.Emit(t.Context(), 1)
			require.ErrorIs(t, err, onceErr)

			err = e.Emit(t.Context(), 2)
			require.NoError(t, err)

			t.Run("Then the listener is called exactly once", func(t *testing.T) {
				assert.Equal(t, 1, calls)
			})
		})
	})
}

func applyTestEventEmitterDispatchMutations(t *testing.T, newEmitter EmitterConstructor) {
	t.Run("Given a listener which unsubscribes itself during dispatch", func(t *testing.T) {
		e := newEmitter()
		calls := 0
		var sub *Subscription[int]
		sub = e.On(func(_ context.Context, _ int) {
			calls++
			sub.Off()
		})

		t.Run("When an event is emitted twice", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 1))
			require.NoError(t, e.Emit(t.Context(), 2))

			t.Run("Then it is called only once", func(t *testing.T) {
				assert.Equal(t, 1, calls)
			})
		})
	})

	t.Run("Given a first listener which unsubscribes a second during dispatch", func(t *testing.T) {
		e := newEmitter()
		firstCalls := 0
		secondCalls := 0
		secondSub := e.On(func(_ context.Context, _ int) {
			secondCalls++
		})
		e.On(func(_ context.Context, _ int) {
			firstCalls++
			secondSub.Off()
		})

		t.Run("When an event is emitted twice", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 1))
			require.NoError(t, e.Emit(t.Context(), 2))

			t.Run("Then the first listener is called both times", func(t *testing.T) {
				assert.Equal(t, 2, firstCalls)
			})
			t.Run("And the second listener is called only once", func(t *testing.T) {
				assert.Equal(t, 1, secondCalls)
			})
		})
	})

	t.Run("Given a listener which adds multiple listeners during dispatch", func(t *testing.T) {
		e := newEmitter()
		firstCalls := 0
		e.On(func(_ context.Context, _ int) {
			firstCalls++
			e.On(func(_ context.Context, event int) {})
			e.On(func(_ context.Context, event int) {})
			e.On(func(_ context.Context, event int) {})
		})

		t.Run("When an event is emitted once", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 1))
			t.Run("Then only the original listener is called", func(t *testing.T) {
				assert.Equal(t, 1, firstCalls)
			})
		})

		t.Run("When emitted a second time", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 2))
			t.Run("Then the original listener is called again", func(t *testing.T) {
				assert.Equal(t, 2, firstCalls)
			})
		})
	})

	t.Run("Given a listener which registers a listener that registers a listener", func(t *testing.T) {
		e := newEmitter()
		outerCalls := 0
		middleCalls := 0
		innerCalls := 0

		e.On(func(_ context.Context, _ int) {
			outerCalls++
			if outerCalls == 1 {
				e.On(func(_ context.Context, _ int) {
					middleCalls++
					if middleCalls == 1 {
						e.On(func(_ context.Context, _ int) {
							innerCalls++
						})
					}
				})
			}
		})

		t.Run("On the first emit only the outer listener fires", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 1))
			assert.Equal(t, 1, outerCalls)
			assert.Equal(t, 0, middleCalls)
			assert.Equal(t, 0, innerCalls)
		})

		t.Run("On the second emit the outer and middle fire", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 2))
			assert.Equal(t, 2, outerCalls)
			assert.Equal(t, 1, middleCalls)
			assert.Equal(t, 0, innerCalls)
		})

		t.Run("On the third emit all three fire", func(t *testing.T) {
			require.NoError(t, e.Emit(t.Context(), 3))
			assert.Equal(t, 3, outerCalls)
			assert.Equal(t, 2, middleCalls)
			assert.Equal(t, 1, innerCalls)
		})
	})
}
