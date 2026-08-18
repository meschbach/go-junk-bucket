package reactors_test

import (
	"context"
	"fmt"
	"time"

	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"github.com/meschbach/go-junk-bucket/pkg/task"
)

func ExampleChannel() {
	ctx := context.Background()

	// Create a channel reactor with a work queue buffer of 10.
	reactor, queue := reactors.NewChannel[int](10)

	// Schedule some work.
	var results []string
	reactor.ScheduleFunc(ctx, func(ctx context.Context) error {
		results = append(results, "first")
		return nil
	})
	reactor.ScheduleFunc(ctx, func(ctx context.Context) error {
		results = append(results, "second")
		return nil
	})

	// Drive the reactor by reading from the queue and ticking.
	event := <-queue
	reactor.Tick(ctx, event, 0)
	event = <-queue
	reactor.Tick(ctx, event, 0)

	fmt.Println(results)

	// Output: [first second]
}

func ExampleRunChannelActor() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// RunChannelActor starts a goroutine that processes events automatically.
	reactor := reactors.RunChannelActor(ctx, 0)

	var result int
	reactor.ScheduleStateFunc(ctx, func(ctx context.Context, state int) error {
		result = state + 42
		return nil
	})

	// Give the goroutine time to process the event.
	time.Sleep(10 * time.Millisecond)
	fmt.Println(result)

	// Output: 42
}

func ExampleTicked() {
	ctx := context.Background()

	// Ticked reactors are manually driven — no goroutine required.
	ticked := &reactors.Ticked[int]{}

	var result int
	ticked.ScheduleStateFunc(ctx, func(ctx context.Context, state int) error {
		result = state * 2
		return nil
	})

	// Tick processes all pending events.
	hasMore, err := ticked.Tick(ctx, 10, 21)
	if err != nil {
		panic(err)
	}

	fmt.Printf("result=%d hasMore=%v\n", result, hasMore)

	// Output: result=42 hasMore=false
}

func ExampleSubmit() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create two reactors: a manually-driven one and an auto-driven one.
	ticked := &reactors.Ticked[int]{}
	autoReactor := reactors.RunChannelActor(ctx, 100)

	// Submit work from ticked to autoReactor. The apply function runs inside
	// autoReactor's boundary, and the result comes back to ticked's boundary.
	tickedCtx := reactors.WithReactor[int](ctx, ticked)
	promise := reactors.Submit[int, int, int](tickedCtx, ticked, autoReactor, func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})

	// Drive ticked to receive the callback. The autoReactor processes the work
	// in its goroutine, then schedules the result back to ticked.
	var result int
	done := make(chan struct{})
	promise.OnCompleted(ctx, func(ctx context.Context, event task.Result[int]) {
		result = event.Output
		close(done)
	})

	// Keep ticking until the result arrives.
	for {
		select {
		case <-done:
			fmt.Println(result)
			return
		default:
			ticked.Tick(ctx, 10, 0)
			time.Sleep(time.Millisecond)
		}
	}

	// Output: 101
}

func Example_context() {
	ctx := context.Background()
	ticked := &reactors.Ticked[string]{}

	// Store the reactor in a child context.
	reactorCtx := reactors.WithReactor(ctx, ticked)

	// Retrieve it later with Maybe (safe — won't panic).
	if boundary, ok := reactors.Maybe[string](reactorCtx); ok {
		fmt.Println("found boundary:", boundary == ticked)
	}

	// Schedule work using the boundary retrieved from context.
	var message string
	reactors.For[string](reactorCtx).ScheduleFunc(reactorCtx, func(ctx context.Context) error {
		// Inside the reactor, we can also get the boundary.
		if b, ok := reactors.Maybe[string](ctx); ok {
			message = fmt.Sprintf("running in reactor: %v", b == ticked)
		}
		return nil
	})

	// Drive the ticked reactor.
	ticked.Tick(ctx, 10, "state")
	fmt.Println(message)

	// Output:
	// found boundary: true
	// running in reactor: true
}
