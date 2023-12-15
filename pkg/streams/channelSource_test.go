package streams

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChannelSource(t *testing.T) {
	t.Run("Implements source", func(t *testing.T) {
		assert.Implements(t, (*Source[int])(nil), NewChannelSource[int](nil))
	})

	t.Run("Given a channel source", func(t *testing.T) {
		scope, scopeDone := context.WithCancel(context.Background())
		t.Cleanup(scopeDone)
		pipe := make(chan int, 3)
		src := NewChannelSource(pipe)

		t.Run("When no value is available", func(t *testing.T) {
			t.Run("Then no results are read", func(t *testing.T) {
				buffer := make([]int, 10)
				count, err := src.ReadSlice(scope, buffer)
				assert.ErrorIs(t, err, UnderRun)
				assert.Equal(t, 0, count)
			})
		})

		t.Run("When values are available", func(t *testing.T) {
			pipe <- 17
			pipe <- 19
			pipe <- 23

			t.Run("Then they are read in order", func(t *testing.T) {
				buffer := make([]int, 10)
				count, err := src.ReadSlice(scope, buffer)
				assert.Equal(t, err, UnderRun)
				if assert.Equal(t, 3, count) {
					assert.Equal(t, 17, buffer[0])
					assert.Equal(t, 19, buffer[1])
					assert.Equal(t, 23, buffer[2])
				}
			})
		})

		t.Run("When the read buffer is smaller than the elements available", func(t *testing.T) {
			pipe <- 29
			pipe <- 31
			pipe <- 41

			t.Run("Then the buffer is fully read", func(t *testing.T) {
				buffer := make([]int, 2)
				count, err := src.ReadSlice(scope, buffer)
				require.NoError(t, err)
				if assert.Equal(t, 2, count) {
					assert.Equal(t, 29, buffer[0])
					assert.Equal(t, 31, buffer[1])
				}
			})
			t.Run("Then a subsequent read will result in the rest of the stream", func(t *testing.T) {
				buffer := make([]int, 2)
				count, err := src.ReadSlice(scope, buffer)
				assert.Equal(t, err, UnderRun)
				if assert.Equal(t, 1, count) {
					assert.Equal(t, 41, buffer[0])
				}
			})
		})

		t.Run("When the source pipe is closed with a partial buffer", func(t *testing.T) {
			pipe <- 43
			close(pipe)

			t.Run("Then the end of the stream is noted", func(t *testing.T) {
				buffer := make([]int, 10)
				count, err := src.ReadSlice(scope, buffer)
				if assert.Equal(t, 1, count) {
					assert.Equal(t, 43, buffer[0])
				}
				assert.ErrorIs(t, err, End)
			})

			t.Run("Then subsequent reads are return nothing from end of stream", func(t *testing.T) {
				buffer := make([]int, 10)
				count, err := src.ReadSlice(scope, buffer)
				assert.Equal(t, 0, count)
				assert.ErrorIs(t, err, End)
			})
		})
	})
}
