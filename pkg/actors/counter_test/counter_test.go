package counter_test

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type counterActor struct {
	count uint
}

func (c *counterActor) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case increment:
		c.count++
	case tell:
		r.Tell(msg.who, c.count)
	}
}

type increment struct{}
type tell struct {
	who actors.Pid
}

func TestSimpleCounter(t *testing.T) {
	t.Parallel()

	sys := local.NewSystem()

	port := sys.NewPort()
	pid := sys.SimpleSpawn(&counterActor{})
	sys.Tell(pid, increment{})
	sys.Tell(pid, tell{who: port.Pid()})
	value, err := port.ReceiveTimeout(100 * time.Millisecond)

	if assert.NoError(t, err) {
		assert.Equal(t, uint(1), value)
	}
}
