package streams

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type write struct {
	value   int
	replyTo chan<- error
}

type writeReply struct {
	problem error
}

func TestChannelConnection(t *testing.T) {
	t.Run("Given a port with a buffered input and output", func(t *testing.T) {
		origin := NewBuffer[int](3, WithBufferTracePrefix[int]("origin"))
		target := NewBuffer[int](3, WithBufferTracePrefix[int]("target"))
		port := NewChannelPort[int](3)

		scope, scopeDone := context.WithTimeout(context.Background(), 1*time.Second)
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
				writeEvents := make(chan write, 32)
				pumper, done := context.WithCancel(scope)
				t.Cleanup(done)
				go func() {
					for {
						select {
						case write := <-writeEvents:
							problem := port.Input.Write(pumper, write.value)
							write.replyTo <- problem
						case <-pumper.Done():
						case e := <-port.Feedback:
							err := port.Input.ConsumeEvent(pumper, e)
							if err != nil && !errors.Is(err, context.Canceled) {
								fmt.Printf("Error after completion: %s\n", err.Error())
								require.NoError(t, err)
							}
						}
					}
				}()
				result := make([]int, 32)
				count, err := ReadAll[int](scope, target, result, func(ctx2 context.Context, count int) (bool, error) {
					if count >= 6 {
						return true, nil
					}
					return false, port.Output.WaitOnEvent(ctx2)
				})
				require.NoError(t, err, "expected to finish full array got %#v", result[:count])
				if assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, result[:count]) {
					assert.Equal(t, 6, count)
				}

				t.Run("And pumping through on the output", func(t *testing.T) {
					_, err := port.Output.PumpTick(scope)
					require.NoError(t, err)

					t.Run("Then it accepts writes again", func(t *testing.T) {
						replyAt := make(chan error)
						writeEvents <- write{
							value:   7,
							replyTo: replyAt,
						}
						assert.NoError(t, <-replyAt, "no writing errors")
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
