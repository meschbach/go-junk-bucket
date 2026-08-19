package streams

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelConnection(t *testing.T) {
	t.Run("Given a port with a buffered input and output", func(t *testing.T) {
		origin := NewBuffer[int](3, WithBufferTracePrefix[int]("origin"))
		target := NewBuffer[int](3, WithBufferTracePrefix[int]("target"))
		port := NewChannelPort[int](3)

		scope, scopeDone := context.WithTimeout(t.Context(), 1*time.Second)
		t.Cleanup(scopeDone)
		_, err := Connect[int](scope, origin, port.Input)
		require.NoError(t, err)
		_, err = Connect[int](scope, port.Output, target)
		require.NoError(t, err)

		t.Run("When overfilling the input side", func(t *testing.T) {
			require.NoError(t, origin.Write(scope, 1))
			require.NoError(t, origin.Write(scope, 2))
			require.NoError(t, origin.Write(scope, 3))
			require.NoError(t, origin.Write(scope, 4))
			require.NoError(t, origin.Write(scope, 5))
			require.ErrorIs(t, origin.Write(scope, 6), Full)

			t.Run("Then it rejects further rights", func(t *testing.T) {
				assert.ErrorIs(t, origin.Write(scope, 7), Overflow)
			})

			t.Run("Then the full output side may be read", func(t *testing.T) {
				// Pump first batch: channel has [1,2,3] from flowing writes
				err = port.Output.Resume(scope)
				require.NoError(t, err)

				// Read from target
				result := make([]int, 32)
				n, err := target.ReadSlice(scope, result)
				require.NoError(t, err)
				require.Equal(t, 3, n)
				require.Equal(t, []int{1, 2, 3}, result[:n])

				// Trigger origin to write [4,5,6] to channel.
				// Resume returns Full when the 3rd write fills the channel; that's expected.
				resumeErr := origin.Resume(scope)
				assert.True(t, errors.Is(resumeErr, nil) || errors.Is(resumeErr, Full),
					"origin.Resume should return nil or Full, got %v", resumeErr)

				// Pump second batch
				err = port.Output.Resume(scope)
				require.NoError(t, err)

				// Read remaining values
				n, err = target.ReadSlice(scope, result[3:])
				require.NoError(t, err)
				require.Equal(t, 3, n)
				require.Equal(t, []int{4, 5, 6}, result[3:6])

				assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result[:6])

				t.Run("And pumping through on the output", func(t *testing.T) {
					_, err := port.Output.PumpTick(scope)
					require.NoError(t, err)

					t.Run("Then it accepts writes again", func(t *testing.T) {
						err := port.Input.Write(scope, 7)
						assert.NoError(t, err, "no writing errors")
					})
				})
			})
		})
	})
}

func ReadAll[T any](ctx context.Context, from Source[T], into []T, waiter func(ctx2 context.Context, count int) (bool, error)) (count int, err error) {
	base := 0
	for {
		count, err := from.ReadSlice(ctx, into[base:])
		if err != nil {
			if errors.Is(err, UnderRun) {
				return base, nil
			}
			return base, err
		}
		base += count
		if count == 0 {
			return base, err
		}
		if base > len(into) {
			return base, nil
		}

		if done, err := waiter(ctx, base); err != nil {
			return base, err
		} else if done {
			return base, nil
		}
	}
}
