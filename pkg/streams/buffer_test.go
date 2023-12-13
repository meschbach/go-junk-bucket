package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBufferStream(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		t.Parallel()
		assert.Implements(t, (*Sink[int])(nil), NewBuffer[int](3))
		assert.Implements(t, (*Source[float32])(nil), NewBuffer[float32](3))
	})

	t.Run("Given a Buffer with a limit of 3", func(t *testing.T) {
		t.Parallel()
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		bufferSize := 3
		s := NewBuffer[int](bufferSize)
		t.Run("When we add up to the limit", func(t *testing.T) {
			assert.NoError(t, s.Write(ctx, 0))
			assert.NoError(t, s.Write(ctx, 1))
			assert.NoError(t, s.Write(ctx, 2))

			t.Run("Then it contains all units", func(t *testing.T) {
				assert.Equal(t, []int{0, 1, 2}, s.Output)
			})

			t.Run("Then refuses all additional items", func(t *testing.T) {
				assert.ErrorIs(t, s.Write(ctx, 3), Full)
			})

			t.Run("And read from", func(t *testing.T) {
				readOut := make([]int, bufferSize)
				count, err := s.ReadSlice(ctx, readOut)
				assert.NoError(t, err)

				t.Run("Then all elements read", func(t *testing.T) {
					assert.Equal(t, bufferSize, count)
				})

				t.Run("Then provides all buffered values", func(t *testing.T) {
					assert.Equal(t, []int{0, 1, 2}, readOut)
				})
			})

			t.Run("And finished", func(t *testing.T) {
				assert.NoError(t, s.Finish(ctx))

				t.Run("Then reading results in End", func(t *testing.T) {
					readOut := make([]int, bufferSize)
					count, err := s.ReadSlice(ctx, readOut)
					assert.Equal(t, 0, count)
					assert.ErrorIs(t, err, End)
				})
				t.Run("Then writing results in Finished", func(t *testing.T) {
					e := s.Write(ctx, 5)
					assert.ErrorIs(t, e, Done)
				})
			})
		})
	})

	t.Run("Given two streams", func(t *testing.T) {
		t.Parallel()
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		source := NewBuffer[int](8)
		sink := NewBuffer[int](8)

		t.Run("When filling source stream", func(t *testing.T) {
			var expectedValues []int
			for i := 0; i < 6; i++ {
				require.NoError(t, source.Write(ctx, i))
				expectedValues = append(expectedValues, i)
			}

			t.Run("And connecting it to the other", func(t *testing.T) {
				connection, err := Connect[int](ctx, source, sink)
				require.NoError(t, err)

				t.Run("Then it moves all buffered elements", func(t *testing.T) {
					assert.Equal(t, expectedValues, sink.Output)
				})

				t.Run("And another element is written", func(t *testing.T) {
					require.NoError(t, source.Write(ctx, 6))
					expectedValues = append(expectedValues, 6)

					t.Run("Then it is passed to the sink", func(t *testing.T) {
						assert.Equal(t, expectedValues, sink.Output)
					})
				})

				t.Run("And the connection is closed", func(t *testing.T) {
					require.NoError(t, connection.Close(ctx))

					t.Run("Then another written element is not transmitted", func(t *testing.T) {
						require.NoError(t, source.Write(ctx, 7))
						values := make([]int, 64)
						count, err := sink.ReadSlice(ctx, values)
						require.NoError(t, err)
						assert.Equal(t, expectedValues, values[0:count])
					})
				})
			})
		})
	})

	t.Run("Given a full sink stream", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 100*time.Millisecond)
		t.Cleanup(done)

		source := NewBuffer[int](1)
		sink := NewBuffer[int](1)
		require.NoError(t, sink.Write(ctx, -1))

		cleanup, err := Connect[int](ctx, source, sink)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, cleanup.Close(ctx))
		})

		t.Run("When resuming the source", func(t *testing.T) {
			require.NoError(t, source.Resume(ctx))

			t.Run("Then no new values are pushed in the buffered sink", func(t *testing.T) {
				assert.Equal(t, []int{-1}, sink.Output)
			})
		})

		t.Run("When writing to the source stream", func(t *testing.T) {
			require.NoError(t, source.Write(ctx, 42))

			t.Run("Then the element is buffered in the source", func(t *testing.T) {
				assert.Equal(t, []int{42}, source.Output)
			})
			t.Run("Then the element is not in the destination", func(t *testing.T) {
				assert.Equal(t, []int{-1}, sink.Output)
			})
		})
	})

	t.Run("Given a stream with unread elements", func(t *testing.T) {
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)
		example := []int{42, 6, 20}

		buffer := NewBuffer[int](4)
		for _, i := range example {
			require.NoError(t, buffer.Write(ctx, i))
		}

		output := make([]int, 1)
		count, err := buffer.ReadSlice(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "single element was read")

		t.Run("When closed for further input", func(t *testing.T) {
			finished := 0
			buffer.SinkEvents().OnFinished.On(func(ctx context.Context, event Sink[int]) {
				finished++
			})
			finishing := 0
			buffer.SinkEvents().OnFinishing.On(func(ctx context.Context, event Sink[int]) {
				finishing++
			})
			end := 0
			buffer.SourceEvents().End.On(func(ctx context.Context, event Source[int]) {
				end++
			})
			require.NoError(t, buffer.Finish(ctx))

			t.Run("Then the finished event is dispatched but not end", func(t *testing.T) {
				assert.Equal(t, 1, finishing, "finishing was dispatched")
				assert.Equal(t, 0, finished, "finished was not dispatched")
				assert.Equal(t, 0, end, "end should not have been dispatched")
			})

			t.Run("Then no further input is accepted", func(t *testing.T) {
				require.ErrorIs(t, buffer.Write(ctx, 99), Done, "Sink may no longer accept input")
			})

			t.Run("Then the elements may still be read", func(t *testing.T) {
				remaining := make([]int, 32)
				count, err := buffer.ReadSlice(ctx, remaining)
				assert.NoError(t, err, "all elements read")
				assert.Equal(t, 2, count, "read correct count")

				t.Run("And an End event is dispatched", func(t *testing.T) {
					assert.Equal(t, 1, finished, "finishing was not dispatched again")
					assert.Equal(t, 1, finished, "finished was dispatched")
					assert.Equal(t, 1, end, "end was expected to be dispatched")
				})
			})
		})
	})
}
