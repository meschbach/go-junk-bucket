package reactors

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func executeReactor(t *testing.T, ctx context.Context, reactor *Channel[int], reactorQueue <-chan ChannelEvent[int]) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-reactorQueue:
			err := reactor.Tick(ctx, e, 0)
			require.NoError(t, err)
		}
	}
}

func TestStreamFromTickToChannel(t *testing.T) {
	testContext, done := context.WithTimeout(context.Background(), 1*time.Second)
	t.Cleanup(done)

	sourceReactor, sourceReactorQueue := NewChannel[int](8)
	go executeReactor(t, testContext, sourceReactor, sourceReactorQueue)

	ticked := &Ticked[int]{}
	tickedInput, sourceReactorOutput, err := StreamBetween[int, int, int](testContext, sourceReactor, ticked)
	require.NoError(t, err)
	//tickedContext := WithReactor[int](testContext, ticked)

	t.Run("When writing a value", func(t *testing.T) {
		hasConsumedEvents := false
		tickedInput.SourceEvents().Data.On(func(ctx context.Context, event int) {
			VerifyWithinBoundary[int](ctx, ticked)
			invokingReactor, has := Maybe[int](ctx)
			if assert.True(t, has, "expected invoking reactor setup properly") {
				assert.Equal(t, ticked, invokingReactor)
			}
			hasConsumedEvents = true
		})
		outputBuffer := streams.NewBuffer[int](32)
		_, err := streams.Connect[int](testContext, tickedInput, outputBuffer)
		require.NoError(t, err, "connecting source pipe to output buffer")

		sourceReactor.ScheduleFunc(testContext, func(ctx context.Context) error {
			VerifyWithinBoundary[int](ctx, sourceReactor)
			fmt.Println("Producing event")
			return sourceReactorOutput.Write(ctx, 42)
		})

		for !hasConsumedEvents {
			hasMore, err := ticked.Tick(testContext, 32, 0)
			require.NoError(t, err)
			assert.False(t, hasMore)
		}

		t.Run("Then the output should receive the value", func(t *testing.T) {
			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(testContext, values)
			require.NoError(t, err)
			if assert.Equal(t, 1, count, "expected to read 1 value") {
				assert.Equal(t, 42, values[0])
			}
		})

		t.Run("And the stream is finished", func(t *testing.T) {
			waiter := &sync.WaitGroup{}
			waiter.Add(1)
			sourceReactor.ScheduleFunc(testContext, func(ctx context.Context) error {
				waiter.Done()
				return sourceReactorOutput.Finish(ctx)
			})
			waiter.Wait()
			remaining, err := ticked.Tick(testContext, 32, 0)
			require.NoError(t, err)
			require.False(t, remaining, "no waiting functions should remain")

			t.Run("Then the stream event is propagated", func(t *testing.T) {
				values := make([]int, 32)
				count, err := outputBuffer.ReadSlice(testContext, values)
				assert.Equal(t, 0, count, "no elements should be read")
				assert.Equal(t, streams.End, err)
			})
		})
	})
}
