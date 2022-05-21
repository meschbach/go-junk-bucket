package counter

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimpleCounter(t *testing.T) {
	t.Parallel()
	root, done := context.WithTimeout(context.Background(), 2*time.Second)
	defer done()

	sys := local.NewSystem()

	port := sys.NewPort()
	pid := sys.Spawn(root, &counterActor{})
	sys.Tell(root, pid, increment{})
	sys.Tell(root, pid, increment{})
	sys.Tell(root, pid, increment{})
	sys.Tell(root, pid, tell{who: port.Pid()})
	value, err := port.ReceiveTimeout(100 * time.Millisecond)

	if assert.NoError(t, err) {
		assert.Equal(t, uint(3), value)
	}
}
