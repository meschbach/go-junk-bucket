package faking

import (
	"fmt"
	"slices"
	"sync"

	"github.com/go-faker/faker/v4"
)

// UniqueUniverse produces a stream of values from Generate such that every value
// returned by Next is distinct from all previously returned values. It is useful
// whenever test or generated data must not contain duplicates (e.g. unique IDs,
// names, or record keys).
//
// For each call to Next, the universe invokes Generate up to Max times until it
// produces a value that has not already been returned. That value is recorded and
// returned. If Generate cannot produce a new value within Max attempts, Next
// panics. Tune Max so your generator has a realistic chance of yielding a fresh
// value on every call.
//
// UniqueUniverse is safe for concurrent use: calls to Next and NextPtr may be
// made from multiple goroutines. Configure the Max and Generate fields before
// sharing the instance; they must not be modified after first use.
//
// Construct one directly for arbitrary types:
//
//	counter := 0
//	domain := &UniqueUniverse[int]{
//		Max: 16,
//		Generate: func() int {
//			counter++
//			return counter
//		},
//	}
//	fmt.Println(domain.Next()) // Outputs 1
//	fmt.Println(domain.Next()) // Outputs 2
//
// For unique words, NewUniqueWords returns a ready-to-use universe.
type UniqueUniverse[T comparable] struct {
	mu sync.Mutex

	generated []T

	// Max is the maximum number of Generate attempts Next makes to find a new
	// value before it panics. It must be greater than zero.
	Max int

	// Generate produces candidate values. It may return duplicates; Next retries
	// until a fresh value is found or Max attempts are exhausted.
	Generate func() T
}

// Next returns a value from Generate that has not been returned before and
// records it so that no future call returns the same value.
//
// Next calls Generate at most Max times per invocation. If every candidate is a
// duplicate, Next panics.
func (u *UniqueUniverse[T]) Next() T {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := 0; i < u.Max; i++ {
		proposed := u.Generate()
		if !slices.Contains(u.generated, proposed) {
			u.generated = append(u.generated, proposed)
			return proposed
		}
	}
	panic(fmt.Sprintf("failed to generate unique value within %d iterations", u.Max))
}

// NextPtr returns a pointer to the unique value produced by Next.
func (u *UniqueUniverse[T]) NextPtr() *T {
	value := u.Next()
	return &value
}

// NewUniqueWords returns a UniqueUniverse that yields unique random words,
// backed by faker.Word and allowing up to 32 attempts per Next call.
func NewUniqueWords() *UniqueUniverse[string] {
	return &UniqueUniverse[string]{
		Generate: func() string {
			return faker.Word()
		},
		Max: 32,
	}
}
