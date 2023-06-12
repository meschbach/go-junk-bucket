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
