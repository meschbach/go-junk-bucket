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
