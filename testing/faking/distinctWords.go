package faking

// DistinctWords returns count unique random words. It panics if it cannot
// generate enough unique words within the internal retry limit.
func DistinctWords(count int) []string {
	u := NewUniqueWords()
	words := make([]string, count)
	for i := 0; i < count; i++ {
		words[i] = u.Next()
	}
	return words
}
