package reactors

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type exampleState struct {
	atomic int
}

func TestTickedReactor(t *testing.T) {
	t.Run("Given a reactor without any scheduled events", func(t *testing.T) {
		r := Ticked[*exampleState]{}
		state := &exampleState{atomic: 0}
		t.Run("Then the rector states so", func(t *testing.T) {
			hasMore, err := r.Tick(context.Background(), 10, state)
			require.NoError(t, err)
			assert.False(t, hasMore)
		})

		t.Run("When an event is scheduled and ran", func(t *testing.T) {
			tickCalled := false
			r.ScheduleFunc(context.Background(), func(ctx context.Context) error {
				tickCalled = true
				return nil
			})
			hasMore, err := r.Tick(context.Background(), 10, state)
			require.NoError(t, err)
			assert.False(t, hasMore)

			t.Run("Then it runs the event", func(t *testing.T) {
				assert.True(t, tickCalled)
			})
		})
	})
}
