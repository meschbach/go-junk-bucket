// Package futures provides a promise mechanism for reactor-based async work.
//
// A [Promise] represents a computation scheduled to run inside a reactor
// boundary. When the computation completes, the result is delivered to
// registered handlers — each executing within their own reactor boundary.
//
// # Core Concepts
//
// Use [PromiseFuncOn] to schedule a function on a reactor and get back a
// Promise. Register handlers with [Promise.HandleFuncOn] or the lower-level
// [Traverse] to receive the result on a different reactor boundary.
//
// # Result Delivery
//
// Results are always delivered through reactor boundaries, preserving
// single-threaded semantics. A handler registered via HandleFuncOn executes
// inside the target reactor, not the source reactor where the promise resolved.
//
// # Blocking
//
// [Promise.Await] provides a blocking wait for promise completion. It creates
// a temporary reactor internally and should be used sparingly — prefer
// HandleFuncOn for non-blocking result handling.
//
// # Relationship to reactors.Submit
//
// The parent [reactors] package provides [reactors.Submit] for cross-reactor
// request/response. Futures operates at a lower level, giving you direct
// control over promise creation and handler registration. Submit is built
// on top of similar mechanics but provides a simpler API for the common case.
package futures
