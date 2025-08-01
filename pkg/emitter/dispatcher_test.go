package emitter

import (
	"context"
	"fmt"
	"testing"
)

func TestDispatcherIsEmitter(t *testing.T) {
	applyTestEventEmitter(t, func() Emitter[int] {
		return NewDispatcher[int]()
	})
}

func ExampleDispatcher() {
	//
	// Given a dispatcher
	//
	dispatcher := NewDispatcher[int]()
	//
	// And a subscription
	//
	dispatcher.On(func(ctx context.Context, event int) {
		fmt.Printf("1st listener: %d\n", event)
	})
	//
	// when we dispatch an event then we should see it on the console
	//
	if err := dispatcher.Emit(context.Background(), 42); err != nil {
		panic(err)
	}
	//
	// When we register a second listener and dispatch an event we should see it twice
	//
	secondSub := dispatcher.On(func(ctx context.Context, event int) {
		fmt.Printf("2nd listener: %d\n", event)
	})
	if err := dispatcher.Emit(context.Background(), 46); err != nil {
		panic(err)
	}
	//
	// When we remove the second subscription, then we should only see a single number
	//
	secondSub.Off()
	if err := dispatcher.Emit(context.Background(), 69); err != nil {
		panic(err)
	}
}
