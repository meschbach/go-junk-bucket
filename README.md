# Junk

A Go library of reusable utilities that simplify common development patterns — event dispatching, functional operations, actor systems, streams, and more.

## Installation

```
go get github.com/meschbach/go-junk-bucket
```

## Packages

| Package | Description |
|---------|-------------|
| [pkg/actors](pkg/actors) | Actor system with spawning, routing, and supervision |
| [pkg/dispatcher](pkg/dispatcher) | Channel-based multi-goroutine event dispatch |
| [pkg/emitter](pkg/emitter) | Callback-based 1:n event dispatch (Observer pattern) |
| [pkg/fx](pkg/fx) | Generic functional utilities for slices |
| [pkg/files](pkg/files) | JSON file reading and writing helpers |
| [pkg/observability](pkg/observability) | OpenTelemetry tracing configuration |
| [pkg/reactors](pkg/reactors) | Event reactor de-multiplexer for single-threaded semantics |
| [pkg/stdgrpc](pkg/stdgrpc) | gRPC testing utilities (in-memory connections) |
| [pkg/stdhttp](pkg/stdhttp) | HTTP response helpers with OTel integration |
| [pkg/streams](pkg/streams) | Object stream Source/Sink semantics |
| [pkg/task](pkg/task) | Promise and Result types for async operations |
| [sub](sub) | Subcommand execution with async stdout/stderr piping |

## Usage

### pkg/fx — Functional Utilities

Generic operations on slices with no side effects. All operations preserve input order.

```go
import "github.com/meschbach/go-junk-bucket/pkg/fx"

// Filter even numbers.
evens := fx.Filter([]int{1, 2, 3, 4, 5, 6}, func(i int) bool {
    return i%2 == 0
})

// Map strings to their lengths.
lengths := fx.Map([]string{"hello", "world"}, func(s string) int {
    return len(s)
})

// Split results into passing and failing groups.
pass, fail := fx.Split(results, func(r Result) bool {
    return r.OK
})
```

### pkg/emitter — Event Dispatching

The Observer pattern with dispatch guarantees: registration-order delivery, error aggregation, and panic containment. Two implementations trade concurrency safety against overhead:

- `Dispatcher` — unsynchronized, suited to a single goroutine (event loop, actor).
- `MutexDispatcher` — serializes registration with a mutex, safe for concurrent use.

```go
import (
    "context"
    "fmt"

    "github.com/meschbach/go-junk-bucket/pkg/emitter"
)

// Single-goroutine emitter.
e := emitter.NewDispatcher[int]()

// Register a listener. The returned Subscription manages its lifecycle.
sub := e.On(func(ctx context.Context, event int) {
    fmt.Println("received:", event)
})

// Broadcast an event to all listeners.
if err := e.Emit(context.Background(), 42); err != nil {
    fmt.Println("emit error:", err)
}

// Unsubscribe when done.
sub.Off()

// For concurrent producers and consumers, use MutexDispatcher.
ce := emitter.NewMutexDispatcher[string]()
ce.On(func(ctx context.Context, event string) {
    fmt.Println("concurrent:", event)
})
```

Listeners that return errors are aggregated; panics are recovered and reported as errors. Use `OnE` or `OnceE` for listeners that can fail.

### pkg/dispatcher — Channel-based Dispatch

A channel-based event dispatch system where listeners receive messages on Go channels. Each listener runs in its own goroutine, providing natural backpressure.

```go
import "github.com/meschbach/go-junk-bucket/pkg/dispatcher"

// Create a dispatcher and listen for messages.
d := dispatcher.NewDispatcher[string]()
ch, done := d.Listen()

// Process messages in a goroutine.
go func() {
    for msg := range ch {
        fmt.Println("got:", msg)
    }
}()

// Broadcast a message to all listeners.
d.Broadcast("hello")

// Clean up: signal the listener to stop, then close the dispatcher.
done()
d.Close()
```

`Consume` provides a convenience wrapper that registers a callback and manages the goroutine lifecycle automatically.

## License

MIT License. See [LICENSE](LICENSE) for details.
