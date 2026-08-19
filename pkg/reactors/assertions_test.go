package reactors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertWithinBoundary[S any](t *testing.T, which context.Context, boundary Boundary[S]) {
	if inBoundary, has := Maybe[S](which); has {
		assert.Equal(t, boundary, inBoundary)
	} else {
		value := which.Value(ContextKey)
		assert.Failf(t, "Boundary non-existent or otherwise wrong", "type %#v, expected type %T", value, *new(S))
	}
}
