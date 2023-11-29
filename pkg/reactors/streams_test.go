package reactors

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStreamThroughBoundary(t *testing.T) {
	t.Run("Given two tick boundaries and a created stream", func(t *testing.T) {
		timedTestContext, onTestDone := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(onTestDone)

		type sourceState struct{}
		type targetState struct{}
		originWell := &Ticked[*sourceState]{}
		originWellContext := WithReactor[*sourceState](timedTestContext, originWell)

		outputWell := &Ticked[*targetState]{}
		source, sink, err := StreamBetween[int, *sourceState, *targetState](timedTestContext, originWell, outputWell)
		outputWellContext := WithReactor[*targetState](timedTestContext, outputWell)
		require.NoError(t, err)

		t.Run("When writing to the sink", func(t *testing.T) {
			exampleValue := 10
			require.NoError(t, sink.Write(originWellContext, exampleValue))
			output := make([]int, 10)

			t.Run("Then the value is not available", func(t *testing.T) {
				count, err := source.ReadSlice(outputWellContext, output)
				require.NoError(t, err)
				assert.Equal(t, 0, count)
			})

			t.Run("And the source boundary is ticked forward", func(t *testing.T) {
				_, err = originWell.Tick(timedTestContext, 10, &sourceState{})
				require.NoError(t, err)

				t.Run("Then the value is not available", func(t *testing.T) {
					count, err := source.ReadSlice(outputWellContext, output)
					require.NoError(t, err)
					assert.Equal(t, 0, count)
				})
			})

			t.Run("And the target boundary is moved forward", func(t *testing.T) {
				_, err = outputWell.Tick(timedTestContext, 10, &targetState{})
				require.NoError(t, err)

				t.Run("Then the value is available", func(t *testing.T) {
					count, err := source.ReadSlice(outputWellContext, output)
					require.NoError(t, err)
					if assert.Equal(t, 1, count) {
						assert.Equal(t, exampleValue, output[0])
					}
				})
			})

			t.Run("And the source stream has reached its end", func(t *testing.T) {
				assert.NoError(t, sink.Finish(originWellContext))

				t.Run("And both wells ticked", func(t *testing.T) {
					_, err = originWell.Tick(timedTestContext, 10, &sourceState{})
					require.NoError(t, err)
					_, err = outputWell.Tick(timedTestContext, 10, &targetState{})
					require.NoError(t, err)

					t.Run("Then the source stream has ended", func(t *testing.T) {
						count, err := source.ReadSlice(outputWellContext, output)
						assert.Equal(t, 0, count)
						assert.ErrorIs(t, err, streams.End)
					})
				})
			})
		})
	})
}
