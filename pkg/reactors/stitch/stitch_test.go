package stitch

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"testing"
)

func TestStitch(t *testing.T) {
	t.Run("Given a new stitch", func(t *testing.T) {
		type state struct {
			i int
		}
		out, _ := New[*state](func(ctx context.Context) (*state, error) {
			return &state{i: 32}, nil
		})
		base := suture.NewSimple("root")
		base.Add(out)
		baseContext, baseDone := context.WithCancel(context.Background())
		go func() {
			err := base.Serve(baseContext)
			if !errors.Is(err, context.Canceled) {
				assert.NoError(t, err)
			}
		}()
		t.Cleanup(baseDone)

		t.Run("When given a succeeding promise", func(t *testing.T) {
			p := Promise(context.Background(), out, func(ctx context.Context, s *state) (int, error) {
				return s.i, nil
			})

			t.Run("Then the promise is resolved", func(t *testing.T) {
				r, err := p.Await(context.Background())
				require.NoError(t, err)
				assert.Equal(t, 32, r.Result)
			})
		})
	})

	t.Run("Given a new stitch immediate mode", func(t *testing.T) {
		initValue := 6
		type state struct {
			i int
		}
		out, _ := New[*state](func(ctx context.Context) (*state, error) {
			return &state{i: initValue}, nil
		})

		t.Run("When given an event", func(t *testing.T) {
			givenState := -1
			out.processor.ScheduleStateFunc(context.Background(), func(ctx context.Context, state *state) error {
				givenState = state.i
				return nil
			})

			t.Run("Then it does not immediately execute", func(t *testing.T) {
				assert.Equal(t, -1, givenState, "init is not executed")
			})

			t.Run("And instructed to consume all events", func(t *testing.T) {
				consumedCount, err := out.ConsumeAll(context.Background(), &ActorState[*state]{})
				require.NoError(t, err)

				t.Run("Then the event is consumed", func(t *testing.T) {
					if assert.Equal(t, 1, consumedCount, "consumed one message") {
						assert.Equal(t, initValue, givenState, "init is not executed")
					}
				})
			})
		})
	})
}
