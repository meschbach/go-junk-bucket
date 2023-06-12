// Package futures provides a promise mechanism against reactors to resolve when event occur in the future.
package futures

type Result[O any] struct {
	Resolved bool
	Result   O
	Problem  error
}
