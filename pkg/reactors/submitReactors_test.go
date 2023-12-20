package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/task"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAsyncWork(t *testing.T) {
	t.Run("Given two reactors", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)
		reactorCtx, timerDone := context.WithTimeout(ctx, 1*time.Second)
		t.Cleanup(timerDone)
		reactorA := &Ticked[int]{}
		reactorB := RunChannelActor(reactorCtx, 42)

		t.Run("When a promise resolves", func(t *testing.T) {
			async := Submit[int, int, int](ctx, reactorA, reactorB, func(boundaryContext context.Context, state int) (int, error) {
				return state + 1, nil
			})
			called := false
			value := -1
			async.OnCompleted(ctx, func(ctx context.Context, event task.Result[int]) {
				called = true
				value = event.Output
			})
			for !called {
				select {
				case <-ctx.Done():
					panic(ctx.Err())
				default:
					if hasMore, err := reactorA.Tick(ctx, 32, 99); err != nil {
						panic(err)
					} else if !hasMore {
						time.Sleep(25 * time.Millisecond)
					}
				}
			}

			t.Run("Then there is no race condition (run with -race -count 10)", func(t *testing.T) {
				assert.True(t, called)
				assert.Equal(t, 43, value)
			})
		})
	})
}
