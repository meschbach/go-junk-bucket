package reactors

import (
	"context"
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
		outputBuffer := streams.NewBuffer[int](32)
		_, err := streams.Connect[int](testContext, tickedInput, outputBuffer)
		require.NoError(t, err, "connecting source pipe to output buffer")

		readOperation := sync.WaitGroup{}
		readOperation.Add(1)
		sourceReactor.ScheduleFunc(testContext, func(ctx context.Context) error {
			VerifyWithinBoundary[int](ctx, sourceReactor)
			defer readOperation.Done()
			return sourceReactorOutput.Write(ctx, 42)
		})

		readOperation.Wait()
		outputSideCount, err := ticked.Tick(testContext, 32, 1)
		require.NoError(t, err)
		assert.False(t, outputSideCount)

		t.Run("Then the output should receive the value", func(t *testing.T) {
			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(testContext, values)
			require.NoError(t, err, "Output: %#v\n", values[0:count])
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

			t.Run("Then the stream finished is propagated", func(t *testing.T) {
				values := make([]int, 32)
				count, err := outputBuffer.ReadSlice(testContext, values)
				assert.Equal(t, 0, count, "no elements should be read")
				assert.Same(t, streams.End, err, "Expected error to state done: %s\n", err.Error())
			})
		})
	})
}
