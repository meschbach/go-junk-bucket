package streams

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferStream(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		t.Parallel()
		assert.Implements(t, (*Sink[int])(nil), NewBuffer[int](3))
		assert.Implements(t, (*Source[float32])(nil), NewBuffer[float32](3))
	})

	t.Run("Given an empty Buffer with a limit of 3", func(t *testing.T) {
		t.Parallel()
		s := NewBuffer[int](3)

		t.Run("When Finish is called on an empty buffer", func(t *testing.T) {
			require.NoError(t, s.Finish(t.Context()))

			t.Run("Then ReadSlice returns End", func(t *testing.T) {
				ctx := t.Context()
				readOut := make([]int, 3)
				count, err := s.ReadSlice(ctx, readOut)
				assert.Equal(t, 0, count, "no elements should be read from an empty finished buffer")
				assert.ErrorIs(t, err, End, "finished empty buffer should report End")
			})
		})
	})

	t.Run("Given a Buffer finishing with empty Output", func(t *testing.T) {
		ctx := t.Context()

		s := NewBuffer[int](3)

		require.NoError(t, s.Write(ctx, 1))
		readOut := make([]int, 1)
		count, err := s.ReadSlice(ctx, readOut)
		require.NoError(t, err)
		require.Equal(t, 1, count)
		require.NoError(t, s.Finish(ctx))

		t.Run("Then ReadSlice returns End on empty finishing buffer", func(t *testing.T) {
			out := make([]int, 3)
			count, err := s.ReadSlice(ctx, out)
			assert.Equal(t, 0, count, "no elements from empty finishing buffer")
			assert.ErrorIs(t, err, End, "empty buffer with writeState=bufferFinishing should report End")
		})
	})

	t.Run("Given a Buffer connected to a ChannelSource via ConnectedPipe", func(t *testing.T) {
		ctx := t.Context()

		t.Run("ReadSlice fires Drained and Finish chain causes ReadSlice to return End", func(t *testing.T) {
			port := NewChannelPort[int](32)
			outputBuffer := NewBuffer[int](32)
			pipe, err := Connect[int](ctx, port.Output, outputBuffer)
			require.NoError(t, err)
			t.Cleanup(func() { _ = pipe.Close(ctx) })

			require.NoError(t, port.Input.Finish(ctx))

			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(ctx, values)
			assert.Equal(t, 0, count, "no elements from empty buffer after finish chain")
			assert.ErrorIs(t, err, End, "ReadSlice should re-check readState after Drained and return End")
		})
	})

	t.Run("Given a Buffer with a limit of 3", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		bufferSize := 3
		s := NewBuffer[int](bufferSize)
		sinkEvents := AttachSinkVerifier[int](s)

		t.Run("When we add up to the limit", func(t *testing.T) {
			assert.NoError(t, s.Write(ctx, 0))
			assert.NoError(t, s.Write(ctx, 1))
			assert.ErrorIs(t, s.Write(ctx, 2), Full)
			assert.Equal(t, 1, sinkEvents.FullCount, "full buffer triggered event")

			t.Run("Then refuses all additional items", func(t *testing.T) {
				assert.ErrorIs(t, s.Write(ctx, 3), Overflow)
			})

			t.Run("Then it contains all units", func(t *testing.T) {
				assert.Equal(t, []int{0, 1, 2}, s.Output)
			})

			t.Run("And read from", func(t *testing.T) {
				preReadAvailableCount := sinkEvents.AvailableCount
				readOut := make([]int, bufferSize)
				count, err := s.ReadSlice(ctx, readOut)
				assert.NoError(t, err, err)

				t.Run("Then provides all buffered values", func(t *testing.T) {
					assert.Equal(t, []int{0, 1, 2}, readOut, "failed to read expected elements")
					assert.Equal(t, bufferSize, count, "failed to read the expected size")
				})

				assert.Less(t, preReadAvailableCount, sinkEvents.AvailableCount, "expected to dispatch an available event")
			})

			t.Run("And finished", func(t *testing.T) {
				assert.NoError(t, s.Finish(ctx))

				t.Run("Then reading results in End", func(t *testing.T) {
					readOut := make([]int, bufferSize)
					count, err := s.ReadSlice(ctx, readOut)
					assert.Equal(t, 0, count, "expected zero elements, got %#v", readOut[:count])
					assert.ErrorIs(t, err, End)
				})
				t.Run("Then writing results in Finished", func(t *testing.T) {
					e := s.Write(ctx, 5)
					assert.ErrorIs(t, e, Done)
				})
			})
		})
	})
}
