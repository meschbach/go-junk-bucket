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
		out := New[*state](func(ctx context.Context) (*state, error) {
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
}
