package streams

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelSinkFinishPropagation(t *testing.T) {
	t.Run("Given a ChannelSink Finish connected to a Buffer", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)

		port := NewChannelPort[int](32)
		outputBuffer := NewBuffer[int](32)
		pipe, err := Connect[int](scope, port.Output, outputBuffer)
		require.NoError(t, err)
		t.Cleanup(func() { _ = pipe.Close(scope) })

		var scheduledPump func()
		port.Input.Push = func(ctx context.Context) error {
			scheduledPump = func() {
				_, _ = port.Output.PumpTick(ctx)
			}
			return nil
		}

		require.NoError(t, port.Input.Finish(scope))

		t.Run("Then Push is called to schedule a pump on the output side", func(t *testing.T) {
			require.NotNil(t, scheduledPump, "Finish should schedule a pump via Push")

			scheduledPump()

			t.Run("And the Buffer reports End after the pump runs", func(t *testing.T) {
				values := make([]int, 32)
				count, err := outputBuffer.ReadSlice(scope, values)
				assert.Equal(t, 0, count, "no elements after stream end")
				assert.ErrorIs(t, err, End, "buffer should report End after Finish propagation")
			})
		})
	})

	t.Run("Given values pumped into a Buffer through a ChannelPort", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)

		port := NewChannelPort[int](32)
		outputBuffer := NewBuffer[int](32)
		pipe, err := Connect[int](scope, port.Output, outputBuffer)
		require.NoError(t, err)
		t.Cleanup(func() { _ = pipe.Close(scope) })

		require.NoError(t, port.Input.Write(scope, 10))
		require.NoError(t, port.Input.Write(scope, 20))
		require.NoError(t, port.Input.Write(scope, 30))

		count, err := port.Output.PumpTick(scope)
		require.NoError(t, err)
		require.Equal(t, 3, count, "all 3 values should be pumped into the buffer")

		t.Run("Then the values are in the Buffer", func(t *testing.T) {
			values := make([]int, 32)
			bufCount, bufErr := outputBuffer.ReadSlice(scope, values)
			require.NoError(t, bufErr)
			require.Equal(t, 3, bufCount)
			assert.Equal(t, []int{10, 20, 30}, values[:bufCount])
		})

		t.Run("When Finish is called on the source side", func(t *testing.T) {
			require.NoError(t, port.Input.Finish(scope))

			t.Run("And the source is pumped", func(t *testing.T) {
				_, pumpErr := port.Output.PumpTick(scope)
				assert.ErrorIs(t, pumpErr, End, "PumpTick should detect closed channel")

				t.Run("Then the Buffer reports End", func(t *testing.T) {
					values := make([]int, 32)
					count, err := outputBuffer.ReadSlice(scope, values)
					assert.Equal(t, 0, count, "no elements after stream end")
					assert.ErrorIs(t, err, End, "buffer should report End after source finish")
				})
			})
		})
	})

	t.Run("Given values sitting in the Go channel when Finish is called", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)

		port := NewChannelPort[int](32)
		outputBuffer := NewBuffer[int](32)
		pipe, err := Connect[int](scope, port.Output, outputBuffer)
		require.NoError(t, err)
		t.Cleanup(func() { _ = pipe.Close(scope) })

		require.NoError(t, port.Input.Write(scope, 100))
		require.NoError(t, port.Input.Write(scope, 200))
		require.NoError(t, port.Input.Write(scope, 300))

		// Finish without pumping — closes channel with values buffered
		require.NoError(t, port.Input.Finish(scope))

		t.Run("Then PumpTick reads all values before End", func(t *testing.T) {
			count, err := port.Output.PumpTick(scope)
			assert.Equal(t, 3, count, "all buffered values should be read from closed channel")
			assert.ErrorIs(t, err, End, "PumpTick should return End after draining closed channel")
		})

		t.Run("Then the Buffer has the values", func(t *testing.T) {
			values := make([]int, 32)
			bufCount, bufErr := outputBuffer.ReadSlice(scope, values)
			require.NoError(t, bufErr)
			require.Equal(t, 3, bufCount)
			assert.Equal(t, []int{100, 200, 300}, values[:bufCount])
		})

		t.Run("Then the Buffer reports End", func(t *testing.T) {
			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(scope, values)
			assert.Equal(t, 0, count, "no elements after stream end")
			assert.ErrorIs(t, err, End, "buffer should report End")
		})
	})
}

func TestChannelSink(t *testing.T) {
	t.Run("Sink Interface Conformance Check", func(t *testing.T) {
		assert.Implements(t, (*Sink[int])(nil), NewChannelSink[int](nil))
	})

	t.Run("Given a channel sink with a custom consumer", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)

		pipe := make(chan int, 3)
		sink := NewChannelSink[int](pipe)
		sinkEvents := AttachSinkVerifier[int](sink)

		t.Run("When writing with available capacity", func(t *testing.T) {
			writtenValue := 5
			require.NoError(t, sink.Write(scope, writtenValue))

			t.Run("Then the value is available on the channel", func(t *testing.T) {
				assert.Equal(t, writtenValue, <-pipe)
			})
			t.Run("Then the stream is not closed", func(t *testing.T) {
				sinkEvents.AssertOpen(t)
			})
		})

		t.Run("When the channel is full", func(t *testing.T) {
			require.NoError(t, sink.Write(scope, 7))
			require.NoError(t, sink.Write(scope, 11))
			require.ErrorIs(t, sink.Write(scope, 13), Full)

			t.Run("Then the stream reports full", func(t *testing.T) {
				assert.ErrorIsf(t, sink.Write(scope, 17), Overflow, "buffer should report full")
			})

			t.Run("Then buffer results in the correct order", func(t *testing.T) {
				assert.Equal(t, 7, <-pipe)
				assert.Equal(t, 11, <-pipe)
				assert.Equal(t, 13, <-pipe)
			})

			t.Run("And the sink is finished", func(t *testing.T) {
				if assert.NoError(t, sink.Finish(scope)) {
					t.Run("Then the pipe is closed", func(t *testing.T) {
						value, ok := <-pipe
						assert.False(t, ok, "Expected closed pipe, got %+v", value)
					})

					t.Run("Then a finishing and finished events are dispatched", func(t *testing.T) {
						assert.Equal(t, 1, sinkEvents.FinishingCount)
						assert.Equal(t, 1, sinkEvents.FinishedCount)
					})
				}
			})
		})
	})

	t.Run("When writing after Finish", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)

		pipe := make(chan int, 3)
		sink := NewChannelSink[int](pipe)
		require.NoError(t, sink.Finish(scope))

		t.Run("Then Write returns Done", func(t *testing.T) {
			assert.ErrorIs(t, sink.Write(scope, 42), Done)
		})
	})

	t.Run("When Finish is called twice", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)

		pipe := make(chan int, 3)
		sink := NewChannelSink[int](pipe)
		require.NoError(t, sink.Finish(scope))

		t.Run("Then the second call returns nil", func(t *testing.T) {
			assert.NoError(t, sink.Finish(scope))
		})

		t.Run("Then the pipe is still closed", func(t *testing.T) {
			value, ok := <-pipe
			assert.False(t, ok, "Expected closed pipe, got %+v", value)
		})
	})
}
