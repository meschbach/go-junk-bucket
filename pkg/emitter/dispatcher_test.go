package emitter

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestDispatcherIsEmitter(t *testing.T) {
	t.Parallel()
	applyTestEventEmitter(t, func() Emitter[int] {
		return NewDispatcher[int]()
	})
}

func ExampleDispatcher() {
	exampleContext := context.Background() //nolint
	//
	// Given a dispatcher
	//
	dispatcher := NewDispatcher[int]()
	//
	// And a subscription
	//
	dispatcher.On(func(_ context.Context, event int) {
		fmt.Printf("1st listener: %d\n", event)
	})
	//
	// when we dispatch an event then we should see it on the console
	//
	if err := dispatcher.Emit(exampleContext, 42); err != nil {
		panic(err)
	}
	//
	// When we register a second listener and dispatch an event we should see it twice
	//
	secondSub := dispatcher.On(func(_ context.Context, event int) {
		fmt.Printf("2nd listener: %d\n", event)
	})
	if err := dispatcher.Emit(exampleContext, 46); err != nil {
		panic(err)
	}
	//
	// When we remove the second subscription, then we should only see a single number
	//
	secondSub.Off()
	if err := dispatcher.Emit(exampleContext, 69); err != nil {
		panic(err)
	}
	// Output:
	// 1st listener: 42
	// 1st listener: 46
	// 2nd listener: 46
	// 1st listener: 69
}

func ExampleDispatcher_OnE() {
	ctx := context.Background() //nolint
	dispatcher := NewDispatcher[int]()

	// OnE registers a listener that can return an error.
	dispatcher.OnE(func(_ context.Context, event int) error {
		fmt.Printf("received: %d\n", event)
		if event < 0 {
			return errors.New("negative event")
		}
		return nil
	})

	// Emit returns nil when the listener succeeds.
	if err := dispatcher.Emit(ctx, 42); err != nil {
		fmt.Println("error:", err)
	}

	// Emit returns the listener's error when it fails.
	if err := dispatcher.Emit(ctx, -1); err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// received: 42
	// received: -1
	// error: negative event
}

func ExampleDispatcher_Once() {
	ctx := context.Background() //nolint
	dispatcher := NewDispatcher[int]()

	// Once registers a listener that fires at most once.
	dispatcher.Once(func(_ context.Context, event int) {
		fmt.Printf("once: %d\n", event)
	})

	dispatcher.Emit(ctx, 1) //nolint
	dispatcher.Emit(ctx, 2) //nolint
	dispatcher.Emit(ctx, 3) //nolint
	// Output:
	// once: 1
}

func ExampleDispatcher_OnceE() {
	ctx := context.Background() //nolint
	dispatcher := NewDispatcher[int]()

	// OnceE registers a one-shot listener that can return an error.
	dispatcher.OnceE(func(_ context.Context, event int) error {
		fmt.Printf("once: %d\n", event)
		return errors.New("done")
	})

	if err := dispatcher.Emit(ctx, 10); err != nil {
		fmt.Println("1st emit error:", err)
	}
	if err := dispatcher.Emit(ctx, 20); err != nil {
		fmt.Println("2nd emit error:", err)
	} else {
		fmt.Println("2nd emit: no error (listener removed)")
	}
	// Output:
	// once: 10
	// 1st emit error: done
	// 2nd emit: no error (listener removed)
}

func ExampleSubscription_Off() {
	ctx := context.Background() //nolint
	dispatcher := NewDispatcher[int]()

	sub := dispatcher.On(func(_ context.Context, event int) {
		fmt.Printf("listener: %d\n", event)
	})

	dispatcher.Emit(ctx, 1) //nolint

	// Unsubscribe via the Subscription.
	sub.Off()

	dispatcher.Emit(ctx, 2) //nolint
	// Output:
	// listener: 1
}
