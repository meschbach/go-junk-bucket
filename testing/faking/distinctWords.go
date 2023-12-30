package faking

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"slices"
)

func DistinctWords(count int) []string {
	attemptCount := 32
	words := make([]string, count)
	for i := 0; i < count; i++ {
		acquired := false
		for attempt := 0; attempt < attemptCount; attempt++ {
			proposed := faker.Word()
			if !slices.Contains(words[0:i], proposed) {
				acquired = true
				words[i] = proposed
				break
			}
		}
		if !acquired {
			panic(fmt.Sprintf("unable to generate unique word in %d attempts", attemptCount))
		}
	}
	return words
}
