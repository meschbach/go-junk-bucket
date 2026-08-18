// Package stitch provides a self-contained reactor unit with managed state lifecycle.
//
// An [Actor] wraps a [reactors.Channel] event loop with lazy-initialized state.
// State is created on first use via an [InitState] function, so Actors don't
// require state to be available before the event loop starts.
//
// # Two Driving Modes
//
// An Actor can be driven in two ways:
//
//   - [Actor.Serve]: runs the event loop in the calling goroutine, blocking until
//     ctx is done. This satisfies the suture.Service interface, making it compatible
//     with suture supervisor trees.
//   - [Actor.ConsumeAll]: processes all pending events immediately and returns.
//     Useful for manual/immediate-mode driving without a dedicated goroutine.
//
// # Convenience Functions
//
// [Submit] schedules a function for execution within an Actor's reactor boundary.
// [Promise] creates a [futures.Promise] that runs within an Actor's reactor boundary,
// returning a future that resolves when the work completes.
//
// # Relationship to the reactors Package
//
// Stitch builds on top of [reactors.Channel] but adds state lifecycle management.
// Use [reactors.Channel] directly when you want full control over state. Use stitch
// when you want an Actor that manages its own state initialization.
package stitch
