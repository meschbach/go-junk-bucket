package streams

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
)

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
}
