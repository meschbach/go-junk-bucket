package fx

// Filter returns a subset of elements for each test which is true
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
