package futures

// Result holds the outcome of a resolved promise.
type Result[O any] struct {
	// Resolved indicates whether the promise has completed.
	Resolved bool

	// Result contains the output value when the promise completed successfully.
	// Only valid when Resolved is true and Problem is nil.
	Result O

	// Problem, when not nil, indicates the promise completed with an error.
	Problem error
}
