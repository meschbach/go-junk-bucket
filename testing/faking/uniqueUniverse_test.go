package faking

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNext_ReturnsUniqueValues(t *testing.T) {
	t.Parallel()
	counter := 0
	u := &UniqueUniverse[int]{
		Generate: func() int {
			counter++
			return counter
		},
		Max: 10,
	}

	seen := make(map[int]bool)
	for range 5 {
		v := u.Next()
		assert.False(t, seen[v], "duplicate value: %d", v)
		seen[v] = true
	}
}

func TestNext_SkipsDuplicates(t *testing.T) {
	t.Parallel()
	calls := 0
	vals := []string{"a", "a", "b", "b", "b", "c"}
	u := &UniqueUniverse[string]{
		Generate: func() string {
			v := vals[calls%len(vals)]
			calls++
			return v
		},
		Max: 10,
	}

	assert.Equal(t, "a", u.Next())
	assert.Equal(t, "b", u.Next())
	assert.Equal(t, "c", u.Next())
}

func TestNext_PanicsWhenExhausted(t *testing.T) {
	t.Parallel()
	u := &UniqueUniverse[string]{
		Generate: func() string { return "same" },
		Max:      3,
	}

	u.Next()
	assert.Panics(t, func() { u.Next() })
}

func TestNext_MaxZero_PanicsImmediately(t *testing.T) {
	t.Parallel()
	u := &UniqueUniverse[int]{
		Generate: func() int { return 1 },
		Max:      0,
	}

	assert.Panics(t, func() { u.Next() })
}

func TestNext_MaxOne_PanicsOnDuplicate(t *testing.T) {
	t.Parallel()
	u := &UniqueUniverse[int]{
		Generate: func() int { return 1 },
		Max:      1,
	}

	u.Next()
	assert.Panics(t, func() { u.Next() })
}

func TestNext_IntType(t *testing.T) {
	t.Parallel()
	i := 0
	u := &UniqueUniverse[int]{
		Generate: func() int {
			i++
			return i * 100
		},
		Max: 5,
	}

	v1 := u.Next()
	v2 := u.Next()
	require.NotEqual(t, v1, v2)
	assert.Equal(t, 100, v1)
	assert.Equal(t, 200, v2)
}

func TestNextPtr_ReturnsUniquePointers(t *testing.T) {
	t.Parallel()
	counter := 0
	u := &UniqueUniverse[string]{
		Generate: func() string {
			counter++
			return fmt.Sprintf("word%d", counter)
		},
		Max: 10,
	}

	p1 := u.NextPtr()
	p2 := u.NextPtr()

	require.NotNil(t, p1)
	require.NotNil(t, p2)
	assert.NotSame(t, p1, p2, "pointers should be distinct addresses")
}

func TestNext_ConcurrentUsage(t *testing.T) {
	t.Parallel()
	counter := 0
	u := &UniqueUniverse[int]{
		Generate: func() int {
			counter++
			return counter
		},
		Max: 10,
	}

	const goroutines = 16
	const callsPerGoroutine = 100
	results := make(chan int, goroutines*callsPerGoroutine)

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range callsPerGoroutine {
				results <- u.Next()
			}
		}()
	}
	wg.Wait()
	close(results)

	seen := make(map[int]bool)
	for v := range results {
		assert.False(t, seen[v], "duplicate value: %d", v)
		seen[v] = true
	}
	assert.Len(t, seen, goroutines*callsPerGoroutine)
}
