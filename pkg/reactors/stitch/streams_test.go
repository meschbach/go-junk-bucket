package stitch

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"testing"
	"time"
)

func TestStreamBetween(t *testing.T) {
	t.Run("Given two stitches and a stream between", func(t *testing.T) {
		originReactor, originProcessor := New[string](func(ctx context.Context) (string, error) {
			return "real", nil
		})
		appropriatedReactor, appropriated := New[string](func(ctx context.Context) (string, error) {
			return "appropriated", nil
		})

		s := suture.NewSimple("real")
		s.Add(originReactor)
		runnerContext, closeProcessors := context.WithCancel(context.Background())
		go func() {
			err := s.Serve(runnerContext)
			if !errors.Is(err, context.Canceled) {
				require.NoError(t, err)
			}
		}()
		t.Cleanup(func() {
			closeProcessors()
		})

		source, sink, err := reactors.StreamBetween[int](context.Background(), originProcessor, appropriated)
		require.NoError(t, err)

		t.Run("When values are written from the origin reactor", func(t *testing.T) {
			exampleValue := 2020
			originProcessor.ScheduleFunc(context.Background(), func(ctx context.Context) error {
				return sink.Write(ctx, exampleValue)
			})

			t.Run("Then the value is produced in the target well", func(t *testing.T) {
				timed, done := context.WithTimeout(context.Background(), 1*time.Second)
				t.Cleanup(done)

				count := 0
				out := make([]int, 10)
				for count == 0 {
					_, err := appropriatedReactor.ConsumeAll(timed, &ActorState[string]{})
					require.NoError(t, err)

					count, err = source.ReadSlice(timed, out)
					assert.ErrorIs(t, err, streams.UnderRun)
				}

				if assert.Equal(t, 1, count) {
					assert.Equal(t, exampleValue, out[0])
				}
			})
		})
	})
}
