// Package reactors provides a single-threaded event processing model for Go.
//
// A reactor is an event demultiplexer that serializes all work within its domain,
// eliminating the need for locks when accessing shared state. Instead of
// coordinating goroutines with mutexes and channels, code schedules functions
// into a reactor and they execute one-at-a-time with access to shared state.
//
// # Core Concepts
//
// The [Boundary] interface is the primary abstraction. It schedules functions
// to run within a reactor that manages state of type S. All scheduled functions
// execute sequentially, so the state can be accessed safely without synchronization.
//
// Two implementations are provided:
//
//   - [Channel]: driven by an external event loop reading from a work queue.
//     Use [RunChannelActor] for a ready-to-go goroutine-based reactor.
//   - [Ticked]: manually driven by calling [Ticked.Tick]. Useful for deterministic
//     testing or embedding in larger event loops.
//
// # Context Integration
//
// Reactors are propagated through context. [WithReactor] stores a boundary in
// a child context, and [For] or [Maybe] retrieve it. Code running inside a reactor
// can use [ScheduleFunc] to schedule additional work without passing the
// boundary explicitly.
//
// # Crossing Boundaries
//
// When work needs to move between reactors, [Submit] provides async
// request/response: it runs a function on a target reactor and delivers
// the result back to a reply-to reactor via a [task.Promise]. [StreamBetween]
// creates a [streams.Source]/[streams.Sink] pair that bridges two reactor boundaries.
//
// # Safety
//
// The "sane" build tag enables runtime verification that operations execute
// within their expected reactor boundary via [VerifyWithinBoundary]. Without this
// tag, VerifyWithinBoundary is a no-op. Use the sane tag during development and
// testing to catch boundary violations early:
//
//	go build -tags sane ./...
//	go test -tags sane ./...
//
// # When to Use Reactors
//
// Reactors are useful when you have asynchronous components that need to share
// state without locks. Common scenarios include:
//
//   - Event-driven architectures where events must be processed in order
//   - State machines with async event sources
//   - Coordinating multiple goroutines that access shared data
//   - Testing async code deterministically with [Ticked]
//
// # When Not to Use Reactors
//
// Reactors serialize all work within a boundary. If your workload benefits from
// parallel execution, consider other concurrency patterns. Reactors are best
// when correctness depends on ordering or when shared state would otherwise
// require complex locking.
package reactors
