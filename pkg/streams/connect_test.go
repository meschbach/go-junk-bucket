package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	t.Run("Given two connected streams with elements passed", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		inputBuffer := NewBuffer[int](8, WithBufferTracePrefix[int]("input"))
		outputBuffer := NewBuffer[int](8, WithBufferTracePrefix[int]("output"))

		_, err := Connect[int](ctx, inputBuffer, outputBuffer)
		require.NoError(t, err)

		t.Run("When filling the input buffer", func(t *testing.T) {
			exampleValues := []int{0, 1, 2, 3, 4, 5, 6, 7}
			for _, v := range exampleValues {
				require.NoError(t, inputBuffer.Write(ctx, v))
			}

			t.Run("Then the values should move to the output buffer", func(t *testing.T) {
				drained := make([]int, 32)
				count, err := ReadAll[int](ctx, outputBuffer, drained, func(ctx2 context.Context, count int) (bool, error) {
					return count != len(exampleValues), nil
				})
				require.NoError(t, err)

				if assert.Equal(t, exampleValues, drained[:count], "expected values are read through") {
					assert.Equal(t, len(exampleValues), count)
				}
			})
		})

		t.Run("When the input buffer with an empty output buffer is closed", func(t *testing.T) {
			require.NoError(t, inputBuffer.Finish(ctx))
			drainedValues := make([]int, 32)
			count, err := outputBuffer.ReadSlice(ctx, drainedValues)
			require.ErrorIs(t, err, End, "end is propagated")
			assert.Equal(t, 0, count, "no elements are read")
		})
	})
}
