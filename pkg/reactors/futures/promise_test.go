package futures

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFutureResolution(t *testing.T) {
	t.Run("Given two reactors", func(t *testing.T) {
		ctx := context.Background()
		initReactor := &reactors.Ticked{}
		secondReactor := &reactors.Ticked{}

		t.Run("When we schedule on the init reactor with a completion notice on the second reactor", func(t *testing.T) {
			result := -1
			PromiseFuncOn(ctx, initReactor, func(ctx context.Context) (int, error) {
				return 0, nil
			}).HandleFuncOn(ctx, secondReactor, func(ctx context.Context, resolved Result[int]) error {
				result = resolved.Result
				return nil
			})

			t.Run("Then no work is done yet", func(t *testing.T) {
				assert.Equal(t, result, -1)
			})

			t.Run("And we run the first reactor", func(t *testing.T) {
				hasMore, err := initReactor.Tick(ctx, 1)
				require.NoError(t, err)
				assert.False(t, hasMore)

				t.Run("Then the second reactor has not received the results", func(t *testing.T) {
					assert.Equal(t, -1, result)
				})

				t.Run("And we run the second reactor", func(t *testing.T) {
					more, err := secondReactor.Tick(ctx, 1)
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
