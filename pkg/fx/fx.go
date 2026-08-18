/*
Package fx provides a small set of functional utilities for operating on slices.

# Motivation

Go's type system makes it easy to write clear, single-purpose loop helpers, but
reimplementing the same patterns—filter, map, split—across projects wastes time
and invites subtle bugs.  This package packages those patterns into tested,
generic functions so callers can compose them without hand-rolling loops.

# Design

Every function in this package is a pure function: it produces no side effects,
does not mutate its inputs, and returns a new slice (or value) derived from the
arguments.  All slice operations preserve the relative order of the input
elements in their outputs.  The functions are generic, so they work with any
element type.

# Usage

Filter, Map, and Split are the core primitives.  They compose naturally:

  - Use Filter to narrow a slice to elements that satisfy a predicate.
  - Use Map to transform every element from one type to another.
  - Use Split to partition a slice into two groups in a single pass.

Identity is a convenience function useful as a default transform in tests.

# Testing

Each function is tested with table-driven subtests covering empty, single-
element, and multi-element inputs.  The Example functions serve as executable
documentation verified by "go test".
*/
package fx

// Filter returns a new slice containing only the elements for which test returns true.  The original slice is left
// unmodified and the relative order of the kept elements is preserved.
func Filter[E any](elements []E, test func(e E) bool) []E {
	out := make([]E, 0)
	for _, e := range elements {
		if test(e) {
			out = append(out, e)
		}
	}
	return out
}

// Split applies all elements e to test.  All e which test returns true are in the first slice, all others are
// in the second slice. Useful for splitting a slice into two buckets to be further operated on.  Filter discards the
// negative results of test which is more efficient if no further processing needs to occur.
func Split[E any](elements []E, test func(e E) bool) ([]E, []E) {
	left := make([]E, 0)
	right := make([]E, 0)
	for _, e := range elements {
		if test(e) {
			left = append(left, e)
		} else {
			right = append(right, e)
		}
	}
	return left, right
}

// Map transforms a slice of inputs through the transform function to result in an equal sized slice of outputs.
func Map[I any, O any](inputs []I, transform func(i I) O) []O {
	inputLength := len(inputs)
	out := make([]O, inputLength, inputLength)
	for index, i := range inputs {
		transformed := transform(i)
		out[index] = transformed
	}
	return out
}

// Identity is an f(x) = x function.  Really useful for testing.
func Identity[V any](input V) V {
	return input
}
