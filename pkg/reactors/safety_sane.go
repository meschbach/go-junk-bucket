//go:build sane

package reactors

import (
	"context"
	"fmt"
)

func VerifyWithinBoundary[S any](ctx context.Context, boundary Boundary[S]) {
	if inBoundary, has := Maybe[S](ctx); has {
		if inBoundary != boundary {
			panic(fmt.Sprintf("expected boundary %s, got %s", boundary, inBoundary))
		}
	} else {
		value := ctx.Value(ContextKey)
		panic(fmt.Sprintf("Boundary non-existent or otherwise wrong type %#v, expected type %T", value, *new(S)))
	}
}
