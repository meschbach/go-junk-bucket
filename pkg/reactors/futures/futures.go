// Package futures provides a promise mechanism against reactors to resolve when event occur in the future.
package futures

type Result[O any] struct {
	Resolved bool
	Result   O
	//Problem, when not nil, indicates the future in question has resolved with an error or has failed
	Problem error
}
