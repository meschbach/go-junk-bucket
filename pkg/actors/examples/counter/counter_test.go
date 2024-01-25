package counter

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleCounter(t *testing.T) {
	t.Parallel()
	root, done := context.WithCancel(context.Background())
	t.Cleanup(done)

	sys := local.NewSystem()

	port := sys.NewPort()
	pid := sys.Spawn(root, &counterActor{})
	sys.Tell(root, pid, increment{})
	sys.Tell(root, pid, increment{})
	sys.Tell(root, pid, increment{})
	sys.Tell(root, pid, tell{who: port.Pid()})
	value, err := port.ReceiveWith(root)

	if assert.NoError(t, err) {
		assert.Equal(t, uint(3), value)
	}
}
