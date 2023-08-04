package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCrossTaskPromise(t *testing.T) {
	//Following test case allows for deterministic validation of correctness and proper flow order.
	t.Run("Given two tick reactors", func(t *testing.T) {
		type aState struct {
			value int
		}
		type bState struct {
			value int
		}

		tickA := &Ticked[*aState]{}
		tickB := &Ticked[*bState]{}

		t.Run("When a promise is made from B", func(t *testing.T) {
			asyncTask := Submit[*bState, *aState, int](context.Background(), tickB, tickA, func(ctx context.Context, state *aState) (int, error) {
				return state.value, nil
			})

			result := -1
			asyncTask.OnCompleted(context.Background(), func(ctx context.Context, event task.Result[int]) {
				result = event.Output
			})

			t.Run("Then the promise is not immediately resolved", func(t *testing.T) {
				assert.Equal(t, -1, result)
			})

			t.Run("And the target reactors are executed", func(t *testing.T) {
				exampleValue := 32
				assertTickedAll(t, tickA, &aState{exampleValue})
				assertTickedAll(t, tickB, &bState{exampleValue * 2})

				t.Run("Then the promise is resolved", func(t *testing.T) {
					assert.Equal(t, exampleValue, result)
				})
			})
		})
	})
}

func assertTickedAll[T any](t *testing.T, reactor *Ticked[T], state T) {
	hasMore := true
	var err error
	for hasMore {
		hasMore, err = reactor.Tick(context.Background(), 10, state)
		require.NoError(t, err)
	}
}
