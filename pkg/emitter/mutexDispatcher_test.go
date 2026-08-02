package emitter

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutexDispatcherIsEmitter(t *testing.T) {
	t.Parallel()
	applyTestEventEmitter(t, func() Emitter[int] {
		return NewMutexDispatcher[int]()
	})
}

func TestMutexDispatcherConcurrentUse(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()

	const producerCount = 4
	const perProducer = 100

	var received atomic.Int32
	subs := make([]*Subscription[int], 0, producerCount)
	for range producerCount {
		subs = append(subs, dispatcher.On(func(_ context.Context, _ int) {
			received.Add(1)
		}))
	}

	ctx := t.Context()
	var wg sync.WaitGroup
	for range producerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range perProducer {
				if err := dispatcher.Emit(ctx, e); err != nil {
					t.Errorf("unexpected emit error: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, producerCount*perProducer*producerCount, int(received.Load()))

	for _, sub := range subs {
		sub.Off()
	}
	require.NoError(t, dispatcher.Emit(ctx, -1))
	assert.Equal(t, producerCount*perProducer*producerCount, int(received.Load()))
}

func ExampleMutexDispatcher() {
	exampleContext := context.Background()
	//
	// Given a concurrency-safe dispatcher
	//
	dispatcher := NewMutexDispatcher[int]()
	//
	// And a subscription
	//
	dispatcher.On(func(_ context.Context, event int) {
		fmt.Printf("listener: %d\n", event)
	})
	//
	// Then an event can be emitted from any goroutine
	//
	if err := dispatcher.Emit(exampleContext, 42); err != nil {
		panic(err)
	}
	// Output:
	// listener: 42
}
