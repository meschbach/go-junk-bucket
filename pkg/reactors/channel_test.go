package reactors

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChannelReactor(t *testing.T) {
	t.Run("Type Compliance", func(t *testing.T) {
		t.Run("Channel as a reactor", func(t *testing.T) {
			assert.Implements(t, (*Boundary[int])(nil), &Channel[int]{})
		})
	})

	t.Run("Given a channel reactor", func(t *testing.T) {
		reactor, queue := NewChannel[int](10)

		t.Run("When a unit is scheduled", func(t *testing.T) {
			called := false
			reactor.ScheduleFunc(context.Background(), func(ctx context.Context) error {
				called = true
				return nil
			})

			t.Run("Then it is not immediately run", func(t *testing.T) {
				assert.False(t, called)
			})

			t.Run("And it is received from the queue and run", func(t *testing.T) {
				op := <-queue
				err := reactor.Tick(context.Background(), op, 0)
				require.NoError(t, err)

				t.Run("Then the unit of work is executed", func(t *testing.T) {
					assert.True(t, called)
				})
			})
		})
	})

	t.Run("Given a channel reactor", func(t *testing.T) {
		reactor, _ := NewChannel[int](10)

		t.Run("When requested to consume all", func(t *testing.T) {
			var invokingContext context.Context
			reactor.ScheduleFunc(context.Background(), func(ctx context.Context) error {
				invokingContext = ctx
				return nil
			})
			count, err := reactor.ConsumeAll(context.Background(), 0)
			assert.Equal(t, 1, count)
			assert.NoError(t, err)

			AssertWithinBoundary[int](t, invokingContext, reactor)
		})
	})
}
