package faking

func DistinctWords(count int) []string {
	u := NewUniqueWords()
	words := make([]string, count)
	for i := 0; i < count; i++ {
		words[i] = u.Next()
	}
	return words
}
