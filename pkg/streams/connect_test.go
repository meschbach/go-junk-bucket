package streams

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		t.Run("When a ChannelSource is connected to a Buffer and the source channel is closed", func(t *testing.T) {
			port := NewChannelPort[int](32)
			outputBuffer := NewBuffer[int](32)

			pipe, err := Connect[int](ctx, port.Output, outputBuffer)
			require.NoError(t, err)
			t.Cleanup(func() { pipe.Close(ctx) })

			// Finish the sink side (closes the Go channel)
			require.NoError(t, port.Input.Finish(ctx))

			// Manually pump the source to detect the closed channel
			count, pumpErr := port.Output.PumpTick(ctx)
			assert.Equal(t, 0, count, "no elements from closed channel")
			assert.ErrorIs(t, pumpErr, End, "PumpTick should detect closed channel")

			t.Run("Then the connected Buffer receives End", func(t *testing.T) {
				values := make([]int, 32)
				bufCount, bufErr := outputBuffer.ReadSlice(ctx, values)
				assert.Equal(t, 0, bufCount, "no elements in finished buffer")
				assert.ErrorIs(t, bufErr, End, "buffer should report End after source finish propagates")
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

	t.Run("ConnectedPipe wiring", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		t.Run("End handler propagates to Buffer", func(t *testing.T) {
			port := NewChannelPort[int](32)
			outputBuffer := NewBuffer[int](32)
			pipe, err := Connect[int](ctx, port.Output, outputBuffer)
			require.NoError(t, err)
			t.Cleanup(func() { _ = pipe.Close(ctx) })

			// Fire End directly on the ChannelSource events
			emitErr := port.Output.SourceEvents().End.Emit(ctx, port.Output)
			require.NoError(t, emitErr)

			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(ctx, values)
			assert.Equal(t, 0, count, "no elements in finished buffer")
			assert.ErrorIs(t, err, End, "End handler should have called Buffer.Finish")
		})

		t.Run("Drained handler resumes source and delivers data", func(t *testing.T) {
			port := NewChannelPort[int](32)
			outputBuffer := NewBuffer[int](32)
			pipe, err := Connect[int](ctx, port.Output, outputBuffer)
			require.NoError(t, err)
			t.Cleanup(func() { _ = pipe.Close(ctx) })

			// Write a value to the Go channel — it sits there until pumped
			require.NoError(t, port.Input.Write(ctx, 42))

			// ReadSlice on empty buffer fires Drained → drain handler calls
			// source.Resume → PumpTick → reads value → writes to Buffer
			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(ctx, values)
			require.NoError(t, err)
			require.Equal(t, 1, count, "Drained chain should have pumped the value into the buffer")
			assert.Equal(t, 42, values[0])
		})

		t.Run("Data handler writes to connected sink", func(t *testing.T) {
			port := NewChannelPort[int](32)
			outputBuffer := NewBuffer[int](32)
			pipe, err := Connect[int](ctx, port.Output, outputBuffer)
			require.NoError(t, err)
			t.Cleanup(func() { _ = pipe.Close(ctx) })

			// Write and pump — Data handler should write to outputBuffer
			require.NoError(t, port.Input.Write(ctx, 99))
			count, err := port.Output.PumpTick(ctx)
			require.NoError(t, err)
			require.Equal(t, 1, count)

			values := make([]int, 32)
			bufCount, bufErr := outputBuffer.ReadSlice(ctx, values)
			require.NoError(t, bufErr)
			require.Equal(t, 1, bufCount)
			assert.Equal(t, 99, values[0])
		})
	})
}
