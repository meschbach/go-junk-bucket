package futures_test

import (
	"context"
	"fmt"
	"time"

	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/reactors/futures"
)

func ExamplePromiseFuncOn() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel-based reactor with initial state 42.
	reactor := reactors.RunChannelActor(ctx, 42)

	// Schedule work that uses the reactor's state to produce output.
	promise := futures.PromiseFuncOn(ctx, reactor, func(ctx context.Context, state int) (string, error) {
		return fmt.Sprintf("computed: %d", state), nil
	})

	// Wait for the result.
	result, err := promise.Await(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Result)

	// Output: computed: 42
}

func ExamplePromise_handleFuncOn() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Two reactors: worker produces, handler receives.
	worker := reactors.RunChannelActor(ctx, 41)
	handler := reactors.RunChannelActor(ctx, 0)

	// Schedule work, result delivered to handler.
	var result string
	promise := futures.PromiseFuncOn(ctx, worker, func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	promise.HandleFuncOn(ctx, handler, func(ctx context.Context, state int, resolved futures.Result[int]) error {
		result = fmt.Sprintf("result: %d", resolved.Result)
		return nil
	})

	// Wait for the handler to process the result.
	time.Sleep(50 * time.Millisecond)

	fmt.Println(result)

	// Output: result: 42
}

func ExampleTraverse() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Producer and consumer are separate reactors.
	producer := reactors.RunChannelActor(ctx, 4)
	consumer := reactors.RunChannelActor(ctx, 0)

	// Producer computes a value.
	promise := futures.PromiseFuncOn(ctx, producer, func(ctx context.Context, state int) (int, error) {
		return state*10 + 2, nil
	})

	// Route the result to consumer using Traverse.
	var delivered bool
	futures.Traverse(ctx, promise, consumer, func(ctx context.Context, state int, resolved futures.Result[int]) error {
		delivered = true
		return nil
	})

	// Wait for the result to be delivered.
	time.Sleep(50 * time.Millisecond)

	fmt.Println(delivered)

	// Output: true
}
