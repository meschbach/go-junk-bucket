//go:build !sane

package reactors

import "context"

// VerifyWithinBoundary is a no-op when built without the "sane" tag.
// Use the "sane" build tag to enable boundary verification during development.
func VerifyWithinBoundary[S any](ctx context.Context, boundary Boundary[S]) {
}
