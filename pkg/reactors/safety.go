//go:build !sane

package reactors

import "context"

// VerifyWithinBoundary will panic if the given context is invoked with the wrong boundary when using the `sane` build
// tag.  Otherwise, this ia no-op
func VerifyWithinBoundary[S any](ctx context.Context, boundary Boundary[S]) {
}
