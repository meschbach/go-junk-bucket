package emitter

import (
	"context"
	"errors"
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

func ExampleMutexDispatcher_OnE() {
	ctx := context.Background() //nolint
	dispatcher := NewMutexDispatcher[int]()

	// OnE registers a concurrency-safe listener that can return an error.
	dispatcher.OnE(func(_ context.Context, event int) error {
		fmt.Printf("received: %d\n", event)
		if event < 0 {
			return errors.New("negative event")
		}
		return nil
	})

	if err := dispatcher.Emit(ctx, 42); err != nil {
		fmt.Println("error:", err)
	}
	if err := dispatcher.Emit(ctx, -1); err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// received: 42
	// received: -1
	// error: negative event
}

func ExampleMutexDispatcher_Once() {
	ctx := context.Background() //nolint
	dispatcher := NewMutexDispatcher[int]()

	// Once registers a concurrency-safe one-shot listener.
	dispatcher.Once(func(_ context.Context, event int) {
		fmt.Printf("once: %d\n", event)
	})

	dispatcher.Emit(ctx, 1) //nolint
	dispatcher.Emit(ctx, 2) //nolint
	// Output:
	// once: 1
}

func TestMutexDispatcherConcurrentOffDuringEmit(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()
	ctx := t.Context()

	var aCalled, bCalled atomic.Int32
	started := make(chan struct{})
	proceed := make(chan struct{})
	var closeOnce sync.Once

	dispatcher.On(func(_ context.Context, _ int) {
		aCalled.Add(1)
		closeOnce.Do(func() { close(started) })
		<-proceed
	})
	bSub := dispatcher.On(func(_ context.Context, _ int) {
		bCalled.Add(1)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-started
		bSub.Off()
		close(proceed)
	}()

	require.NoError(t, dispatcher.Emit(ctx, 1))
	wg.Wait()

	assert.Equal(t, int32(1), aCalled.Load(), "A should be called once")
	assert.Equal(t, int32(1), bCalled.Load(), "B should be called from the snapshot despite Off during Emit")

	require.NoError(t, dispatcher.Emit(ctx, 2))
	assert.Equal(t, int32(2), aCalled.Load(), "A should be called on second emit")
	assert.Equal(t, int32(1), bCalled.Load(), "B should not be called after removal")
}

func TestMutexDispatcherConcurrentOnDuringEmit(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()
	ctx := t.Context()

	var aCalled, bCalled atomic.Int32
	started := make(chan struct{})
	proceed := make(chan struct{})
	var closeOnce sync.Once

	dispatcher.On(func(_ context.Context, _ int) {
		aCalled.Add(1)
		closeOnce.Do(func() { close(started) })
		<-proceed
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-started
		dispatcher.On(func(_ context.Context, _ int) {
			bCalled.Add(1)
		})
		close(proceed)
	}()

	require.NoError(t, dispatcher.Emit(ctx, 1))
	wg.Wait()

	assert.Equal(t, int32(1), aCalled.Load(), "A should be called")
	assert.Equal(t, int32(0), bCalled.Load(), "B should not be called during the emit it was registered in")

	require.NoError(t, dispatcher.Emit(ctx, 2))
	assert.Equal(t, int32(2), aCalled.Load())
	assert.Equal(t, int32(1), bCalled.Load(), "B should be called on the second emit")
}

func TestMutexDispatcherConcurrentOnAndOffDuringEmit(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()
	ctx := t.Context()

	var aCalls, bCalls, cCalls atomic.Int32
	started := make(chan struct{})
	proceed := make(chan struct{})
	var closeOnce sync.Once

	dispatcher.On(func(_ context.Context, _ int) {
		aCalls.Add(1)
		closeOnce.Do(func() { close(started) })
		<-proceed
	})
	dispatcher.On(func(_ context.Context, _ int) {
		bCalls.Add(1)
	})
	cSub := dispatcher.On(func(_ context.Context, _ int) {
		cCalls.Add(1)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-started
		dispatcher.On(func(_ context.Context, _ int) {})
		cSub.Off()
		close(proceed)
	}()

	require.NoError(t, dispatcher.Emit(ctx, 1))
	wg.Wait()

	assert.Equal(t, int32(1), aCalls.Load(), "A should be called")
	assert.Equal(t, int32(1), bCalls.Load(), "B should be called from snapshot")
	assert.Equal(t, int32(1), cCalls.Load(), "C should be called from snapshot despite Off during Emit")

	require.NoError(t, dispatcher.Emit(ctx, 2))
	assert.Equal(t, int32(2), aCalls.Load())
	assert.Equal(t, int32(2), bCalls.Load())
	assert.Equal(t, int32(1), cCalls.Load(), "C should not be called after removal")
}

func TestMutexDispatcherConcurrentStress(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()
	ctx := t.Context()

	const emitters = 8
	const subscribers = 8
	const eventsPerEmitter = 200

	var totalDeliveries atomic.Int64
	subs := make([]*Subscription[int], subscribers)
	for i := range subscribers {
		subs[i] = dispatcher.On(func(_ context.Context, _ int) {
			totalDeliveries.Add(1)
		})
	}

	var wg sync.WaitGroup

	for range emitters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range eventsPerEmitter {
				if err := dispatcher.Emit(ctx, e); err != nil {
					t.Errorf("unexpected emit error: %v", err)
					return
				}
			}
		}()
	}

	var mutateWg sync.WaitGroup
	mutateWg.Add(1)
	go func() {
		defer mutateWg.Done()
		for i, sub := range subs {
			if i%2 == 0 {
				sub.Off()
			}
		}
	}()

	mutateWg.Add(1)
	go func() {
		defer mutateWg.Done()
		for range subscribers {
			dispatcher.On(func(_ context.Context, _ int) {
				totalDeliveries.Add(1)
			})
		}
	}()

	wg.Wait()
	mutateWg.Wait()

	assert.Greater(t, totalDeliveries.Load(), int64(0), "should have at least some deliveries")
}

func TestMutexDispatcherSelfUnsubscribeConcurrent(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()
	ctx := t.Context()

	const emits = 50
	var aCalls, bCalls atomic.Int32

	var aSub *Subscription[int]
	aSub = dispatcher.On(func(_ context.Context, _ int) {
		aCalls.Add(1)
		aSub.Off()
	})
	dispatcher.On(func(_ context.Context, _ int) {
		bCalls.Add(1)
	})

	var wg sync.WaitGroup
	for range emits {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := dispatcher.Emit(ctx, 1); err != nil {
				t.Errorf("unexpected emit error: %v", err)
			}
		}()
	}
	wg.Wait()

	assert.GreaterOrEqual(t, aCalls.Load(), int32(1), "self-unsubscribing listener should fire at least once")
	assert.LessOrEqual(t, aCalls.Load(), int32(emits), "self-unsubscribing listener should not fire more than the number of emits")
	assert.GreaterOrEqual(t, bCalls.Load(), int32(1), "the second listener should fire at least once")
	assert.LessOrEqual(t, bCalls.Load(), int32(emits), "the second listener should not fire more than the number of emits")
}

func TestMutexDispatcherConcurrentOnceE(t *testing.T) {
	t.Parallel()
	dispatcher := NewMutexDispatcher[int]()
	ctx := t.Context()

	const onceListeners = 16
	const emits = 100
	calls := make([]atomic.Int32, onceListeners)

	for i := range onceListeners {
		i := i
		dispatcher.OnceE(func(_ context.Context, _ int) error {
			calls[i].Add(1)
			return nil
		})
	}

	var wg sync.WaitGroup
	for range emits {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := dispatcher.Emit(ctx, 1); err != nil {
				t.Errorf("unexpected emit error: %v", err)
			}
		}()
	}
	wg.Wait()

	for i, c := range calls {
		assert.GreaterOrEqual(t, c.Load(), int32(1), "OnceE listener %d should have been called at least once", i)
	}
}
