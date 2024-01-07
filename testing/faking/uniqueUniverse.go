package faking

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"slices"
)

type UniqueUniverse[T comparable] struct {
	generated []T
	//Max is the total number of attempts to make for each Next
	Max      int
	Generate func() T
}

func (u *UniqueUniverse[T]) Next() T {
	for i := 0; i < u.Max; i++ {
		proposed := u.Generate()
		if !slices.Contains(u.generated, proposed) {
			u.generated = append(u.generated, proposed)
			return proposed
		}
	}
	panic(fmt.Sprintf("failed to generate unique value within %d iterations", u.Max))
}

func NewUniqueWords() *UniqueUniverse[string] {
	return &UniqueUniverse[string]{
		Generate: func() string {
			return faker.Word()
		},
		Max: 32,
	}
}
