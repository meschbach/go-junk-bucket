package streams

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	t.Run("Given two connected streams with elements passed", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		source := NewBuffer[int](8)
		sink := NewBuffer[int](8)

		_, err := Connect[int](ctx, source, sink)
		require.NoError(t, err)
		exampleValues := []int{0, 1, 2, 3, 4, 5, 6, 7}
		for _, v := range exampleValues {
			require.NoError(t, source.Write(ctx, v))
		}

		t.Run("When finishing the source", func(t *testing.T) {
			require.NoError(t, source.Finish(ctx))

			t.Run("Then the sink receives all elements", func(t *testing.T) {
				assert.Equal(t, exampleValues, sink.Output)
			})

			t.Run("Then the sink is also finished", func(t *testing.T) {
				assert.Equal(t, bufferFinishing, sink.writeState, fmt.Sprintf("target write state should be finishing, got %s", sink.writeState))
			})
		})
	})
}
