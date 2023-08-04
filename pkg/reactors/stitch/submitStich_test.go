package stitch

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"testing"
	"time"
)

func TestSubmitStitchInteraction(t *testing.T) {
	t.Parallel()
	t.Run("Given an autonomous stitch unit and an appropriated unit", func(t *testing.T) {
		t.Parallel()
		type appropriatedState struct{}
		appropriatedReactor, appropriated := New[*appropriatedState](func(ctx context.Context) (*appropriatedState, error) {
			return nil, nil
		})
		exampleValue := 2006
		type realState struct{ value int }
		realProcessorReactor, realProcessor := New[*realState](func(ctx context.Context) (*realState, error) {
			return &realState{exampleValue}, nil
		})

		s := suture.NewSimple("real")
		s.Add(realProcessorReactor)
		runnerContext, closeProcessors := context.WithCancel(context.Background())
		go func() {
			err := s.Serve(runnerContext)
			if !errors.Is(err, context.Canceled) {
				require.NoError(t, err)
			}
		}()
		t.Cleanup(func() {
			closeProcessors()
		})

		t.Run("When a promise is made from B", func(t *testing.T) {
			asyncTask := reactors.Submit[*appropriatedState, *realState, int](context.Background(), appropriated, realProcessor, func(ctx context.Context, state *realState) (int, error) {
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
				ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
				t.Cleanup(done)
				state := &ActorState[*appropriatedState]{}
				for {
					_, err := appropriatedReactor.ConsumeAll(ctx, state)
					if errors.Is(err, context.DeadlineExceeded) {
						break
					}
					require.NoError(t, err)
					if result == exampleValue {
						break
					}
				}

				t.Run("Then the promise is resolved", func(t *testing.T) {
					assert.Equal(t, exampleValue, result)
				})
			})
		})
	})
}
