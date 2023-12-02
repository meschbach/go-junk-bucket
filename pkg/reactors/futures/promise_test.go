package futures

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type tickedExampleState struct {
}

func TestFutureResolutionWithTicked(t *testing.T) {
	t.Run("Given two reactors", func(t *testing.T) {
		ctx := context.Background()
		initReactor := &reactors.Ticked[*tickedExampleState]{}
		secondReactor := &reactors.Ticked[*tickedExampleState]{}

		t.Run("When we schedule on the init reactor with a completion notice on the second reactor", func(t *testing.T) {
			result := -1
			PromiseFuncOn[*tickedExampleState, int](ctx, initReactor, func(ctx context.Context, state *tickedExampleState) (int, error) {
				return 0, nil
			}).HandleFuncOn(ctx, secondReactor, func(ctx context.Context, state *tickedExampleState, resolved Result[int]) error {
				result = resolved.Result
				return nil
			})

			t.Run("Then no work is done yet", func(t *testing.T) {
				assert.Equal(t, result, -1)
			})

			t.Run("And we run the first reactor", func(t *testing.T) {
				hasMore, err := initReactor.Tick(ctx, 1, nil)
				require.NoError(t, err)
				assert.False(t, hasMore)

				t.Run("Then the second reactor has not received the results", func(t *testing.T) {
					assert.Equal(t, -1, result)
				})

				t.Run("And we run the second reactor", func(t *testing.T) {
					more, err := secondReactor.Tick(ctx, 1, nil)
					require.NoError(t, err)
					assert.False(t, more)

					t.Run("Then the expected result is set", func(t *testing.T) {
						assert.Equal(t, 0, result)
					})
				})
			})
		})
	})
}

func TestFutureResolutionWithChannelsAndTicked(t *testing.T) {
	t.Run("Given two reactors", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)
		initReactor := &reactors.Ticked[*tickedExampleState]{}
		secondReactor, input := reactors.NewChannel[*tickedExampleState](10)
		go func() {
			for {
				select {
				case e := <-input:
					if err := secondReactor.Tick(ctx, e, nil); err != nil {
						panic(err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		t.Run("When we schedule on the init reactor with a completion notice on the second reactor", func(t *testing.T) {
			result := -1
			PromiseFuncOn[*tickedExampleState, int](ctx, secondReactor, func(ctx context.Context, state *tickedExampleState) (int, error) {
				return 0, nil
			}).HandleFuncOn(ctx, initReactor, func(ctx context.Context, state *tickedExampleState, resolved Result[int]) error {
				result = resolved.Result
				return nil
			})

			t.Run("Then no work is done yet", func(t *testing.T) {
				assert.Equal(t, result, -1)
			})

			t.Run("And we run the first reactor", func(t *testing.T) {
				for result == -1 {
					hasMore, err := initReactor.Tick(ctx, 10, nil)
					require.NoError(t, err)
					assert.False(t, hasMore)
				}

				t.Run("And we run the second reactor", func(t *testing.T) {
					t.Run("Then the expected result is set", func(t *testing.T) {
						assert.Equal(t, 0, result)
					})
				})
			})
		})
	})
}
