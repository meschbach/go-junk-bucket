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

func TestStreamFinishEndToEnd(t *testing.T) {
	t.Run("Given a StreamBetween bridging a Channel source to a Ticked output with a connected Buffer", func(t *testing.T) {
		testContext, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		sourceReactor := RunChannelActor[int](testContext, 0)
		ticked := &Ticked[int]{}
		tickedInput, sourceReactorOutput, err := StreamBetween[int, int, int](testContext, sourceReactor, ticked)
		require.NoError(t, err)

		outputBuffer := streams.NewBuffer[int](32)
		_, err = streams.Connect[int](testContext, tickedInput, outputBuffer)
		require.NoError(t, err)

		t.Run("When Finish is called on the source output", func(t *testing.T) {
			waiter := &sync.WaitGroup{}
			waiter.Add(1)
			sourceReactor.ScheduleFunc(testContext, func(ctx context.Context) error {
				waiter.Done()
				return sourceReactorOutput.Finish(ctx)
			})
			waiter.Wait()

			t.Run("And the output side is ticked to process the Finish", func(t *testing.T) {
				remaining, err := ticked.Tick(testContext, 32, 1)
				assert.ErrorIs(t, err, streams.End, "Tick should propagate End from PumpTick")
				require.False(t, remaining, "no pending tasks should remain after tick")

				t.Run("Then the Buffer reports End", func(t *testing.T) {
					values := make([]int, 32)
					count, err := outputBuffer.ReadSlice(testContext, values)
					assert.Equal(t, 0, count, "no elements should be read from a finished stream")
					assert.ErrorIs(t, err, streams.End, "finished stream should propagate End to the buffer")
				})
			})
		})
	})

	t.Run("Given a StreamBetween with values flowing before Finish", func(t *testing.T) {
		testContext, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)

		sourceReactor := RunChannelActor[int](testContext, 0)
		ticked := &Ticked[int]{}
		tickedInput, sourceReactorOutput, err := StreamBetween[int, int, int](testContext, sourceReactor, ticked)
		require.NoError(t, err)

		outputBuffer := streams.NewBuffer[int](32)
		_, err = streams.Connect[int](testContext, tickedInput, outputBuffer)
		require.NoError(t, err)

		t.Run("When a value is written and ticked through", func(t *testing.T) {
			waiter := &sync.WaitGroup{}
			waiter.Add(1)
			sourceReactor.ScheduleFunc(testContext, func(ctx context.Context) error {
				defer waiter.Done()
				return sourceReactorOutput.Write(ctx, 42)
			})
			waiter.Wait()

			remaining, err := ticked.Tick(testContext, 32, 1)
			require.NoError(t, err)
			require.False(t, remaining)

			values := make([]int, 32)
			count, err := outputBuffer.ReadSlice(testContext, values)
			require.NoError(t, err)
			require.Equal(t, 1, count)
			require.Equal(t, 42, values[0])

			t.Run("And then Finish is called and the output side is ticked", func(t *testing.T) {
				waiter2 := &sync.WaitGroup{}
				waiter2.Add(1)
				sourceReactor.ScheduleFunc(testContext, func(ctx context.Context) error {
					waiter2.Done()
					return sourceReactorOutput.Finish(ctx)
				})
				waiter2.Wait()

				remaining, err := ticked.Tick(testContext, 32, 1)
				assert.ErrorIs(t, err, streams.End, "Tick should propagate End from PumpTick")
				require.False(t, remaining)

				t.Run("Then the Buffer reports End", func(t *testing.T) {
					count, err := outputBuffer.ReadSlice(testContext, values)
					assert.Equal(t, 0, count, "no elements after stream end")
					assert.ErrorIs(t, err, streams.End, "End should propagate after value delivery and Finish")
				})
			})
		})
	})
}

func TestStreamThroughBoundary(t *testing.T) {
	t.Run("Given two tick boundaries and a created stream", func(t *testing.T) {
		timedTestContext, onTestDone := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(onTestDone)

		type sourceState struct{}
		type targetState struct{}
		originWell := &Ticked[*sourceState]{}
		originWellContext := WithReactor[*sourceState](timedTestContext, originWell)

		outputWell := &Ticked[*targetState]{}
		source, sink, err := StreamBetween[int, *sourceState, *targetState](timedTestContext, originWell, outputWell)
		outputWellContext := WithReactor[*targetState](timedTestContext, outputWell)
		require.NoError(t, err)

		t.Run("When writing to the sink", func(t *testing.T) {
			exampleValue := 10
			require.NoError(t, sink.Write(originWellContext, exampleValue))
			output := make([]int, 10)

			t.Run("Then the value is available by read", func(t *testing.T) {
				count, err := source.ReadSlice(outputWellContext, output)
				assert.ErrorIs(t, err, streams.UnderRun)
				if assert.Equal(t, 1, count, "Then the count is correct") {
					assert.Equal(t, exampleValue, output[0])
				}
			})

			t.Run("And the source stream has reached its end", func(t *testing.T) {
				assert.NoError(t, sink.Finish(originWellContext))

				t.Run("And both wells ticked", func(t *testing.T) {
					_, err = originWell.Tick(timedTestContext, 10, &sourceState{})
					require.NoError(t, err)
					_, err = outputWell.Tick(timedTestContext, 10, &targetState{})
					require.NoError(t, err)

					t.Run("Then the source stream has ended", func(t *testing.T) {
						count, err := source.ReadSlice(outputWellContext, output)
						assert.Equal(t, 0, count)
						assert.ErrorIs(t, err, streams.End)
					})
				})
			})
		})
	})
}
