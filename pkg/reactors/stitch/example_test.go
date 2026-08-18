package stitch_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/meschbach/go-junk-bucket/pkg/reactors/stitch"
	"github.com/thejerf/suture/v4"
)

func ExampleNew() {
	ctx := context.Background()

	// Create an actor with lazy-initialized state.
	actor, boundary := stitch.New(func(ctx context.Context) (string, error) {
		return "initialized", nil
	})

	// Schedule work using the boundary.
	boundary.ScheduleStateFunc(ctx, func(ctx context.Context, state string) error {
		fmt.Println("state:", state)
		return nil
	})

	// Drive the actor with ConsumeAll (immediate mode, no goroutine needed).
	state := &stitch.ActorState[string]{}
	count, err := actor.ConsumeAll(ctx, state)
	if err != nil {
		panic(err)
	}
	fmt.Println("consumed:", count)

	// Output:
	// state: initialized
	// consumed: 1
}

func ExampleActor_Submit() {
	ctx := context.Background()

	actor, _ := stitch.New(func(ctx context.Context) (int, error) {
		return 100, nil
	})

	// Submit a function to the actor.
	var result int
	actor.Submit(ctx, func(ctx context.Context, state int) error {
		result = state + 1
		return nil
	})

	// Drive with ConsumeAll.
	actor.ConsumeAll(ctx, &stitch.ActorState[int]{})

	fmt.Println(result)

	// Output: 101
}

func ExampleActor_Serve() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actor, _ := stitch.New(func(ctx context.Context) (int, error) {
		return 42, nil
	})

	// Wrap in a suture supervisor tree.
	supervisor := suture.NewSimple("root")
	supervisor.Add(actor)

	go func() {
		err := supervisor.Serve(ctx)
		if !errors.Is(err, context.Canceled) {
			fmt.Println("supervisor error:", err)
		}
	}()

	// Submit work and await a promise.
	p := stitch.Promise(ctx, actor, func(ctx context.Context, state int) (string, error) {
		return fmt.Sprintf("result: %d", state), nil
	})

	result, err := p.Await(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Result)

	// Output: result: 42
}
