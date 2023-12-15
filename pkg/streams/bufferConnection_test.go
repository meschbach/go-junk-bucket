package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// Tests Buffer to Buffer semantics with the Connect family of controls.
func TestBufferConnection(t *testing.T) {
	t.Run("Given a source stream with existing elements connecting to an output buffer", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		examples := []int{1941, 1942, 1943, 1944, 1945}

		//Construct and pre-load input buffer
		inputBuffer := NewBuffer[int](6)
		for _, e := range examples {
			assert.NoError(t, inputBuffer.Write(ctx, e))
		}
		//inputSinkObserver := AttachSinkVerifier[int](inputBuffer)

		//Establish watches
		outputBuffer := NewBuffer[int](8)
		outputSinkEventsWatcher := AttachSinkVerifier[int](outputBuffer)

		_, err := Connect[int](ctx, inputBuffer, outputBuffer)
		assert.NoError(t, err)

		t.Run("Then the elements are piped through", func(t *testing.T) {
			capture := make([]int, 32)
			count, err := outputBuffer.ReadSlice(ctx, capture)
			require.NoError(t, err)
			if assert.Equal(t, examples, capture[:len(examples)], "expected all elements to be copied") {
				assert.Equal(t, len(examples), count, "count and copied data mismatch")
			}
		})

		t.Run("With no additional elements", func(t *testing.T) {
			drainCountBeforeRead := outputSinkEventsWatcher.DrainedCount
			capture := make([]int, 32)
			count, err := outputBuffer.ReadSlice(ctx, capture)

			t.Run("Then there is a buffer under run", func(t *testing.T) {
				assert.ErrorIs(t, err, UnderRun)
				assert.Equal(t, 0, count, "expected no elements, got %#v", capture[:count])
			})

			t.Run("Then a drain call is issued", func(t *testing.T) {
				assert.Less(t, drainCountBeforeRead, outputSinkEventsWatcher.DrainedCount)
			})
		})

		t.Run("When new elements are written", func(t *testing.T) {
			newElements := []int{48, 11, 34, 22}
			for _, e := range newElements {
				require.NoError(t, inputBuffer.Write(ctx, e))
			}

			t.Run("Then the elements are available on the output buffer", func(t *testing.T) {
				capture := make([]int, 32)
				count, err := outputBuffer.ReadSlice(ctx, capture)
				assert.NoError(t, err)
				assert.Equal(t, newElements, capture[:count])
			})
		})
	})
}
