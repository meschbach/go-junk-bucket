package reactive

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLimitedSliceAccumulator(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		t.Parallel()
		//t.Skip()
		assert.Implements(t, (*Sink[int])(nil), NewBuffer[int](3))
		assert.Implements(t, (*Source[float32])(nil), NewBuffer[float32](3))
	})

	t.Run("Given a Limited accumulator of 3", func(t *testing.T) {
		t.Parallel()
		//t.Skip()
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
		//t.Skip()
		ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		source := NewBuffer[int](8)
		sink := NewBuffer[int](8)

		t.Run("When filling one stream", func(t *testing.T) {
			for i := 0; i < 6; i++ {
				require.NoError(t, source.Write(ctx, i))
			}

			t.Run("And connecting it to the other", func(t *testing.T) {
				connection, err := Connect[int](ctx, source, sink)
				require.NoError(t, err)

				t.Run("Then it moves all buffered elements", func(t *testing.T) {
					assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, sink.Output)
				})

				t.Run("And another element is written", func(t *testing.T) {
					require.NoError(t, source.Write(ctx, 6))

					t.Run("Then it is passed to the sink", func(t *testing.T) {
						assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6}, sink.Output)
					})
				})

				t.Run("And is closed", func(t *testing.T) {
					require.NoError(t, connection.Close())

					t.Run("Then another written element is not transmitted", func(t *testing.T) {
						require.NoError(t, source.Write(ctx, 7))
						assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6}, sink.Output)
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
			require.NoError(t, cleanup.Close())
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
}
