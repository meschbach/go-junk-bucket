package task

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	"testing"
)

func TestAsyncWork(t *testing.T) {
	t.Run("Given two reactors", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		reactorA := reactors.NewChannel(10)

		t.Run("When a promise resolves", func(t *testing.T) {
			t.Run("Then there is no race condition (run with -race -count 10)", func(t *testing.T) {

			})
		})
	})
}
