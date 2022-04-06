package simple_test

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type CounterActor struct {
	count int
}

type Tell struct {
	mailbox *local.LocalMailbox
}

func (c *CounterActor) OnMessage(m any) {
	switch msg := m.(type) {
	case Tell:
		msg.mailbox.Tell(c.count)
	case *Tell:
		msg.mailbox.Tell(c.count)
	case actors.SelfStarted:
	default:
		c.count++
	}
}

func TestSimpleAck(t *testing.T) {
	t.Parallel()

	sys := local.NewLocalActorSystem()
	counter := &CounterActor{count: 0}
	ref := sys.Spawn("test", counter)
	ref.Tell(nil)
	sys.Shutdown()

	assert.Equal(t, 1, counter.count)
}

func TestMultipleAcksSync(t *testing.T) {
	t.Parallel()

	sys := local.NewLocalActorSystem()
	externalMailbox := sys.ExternalMailbox()
	ref := sys.Spawn("test", &CounterActor{count: 0})
	ref.Tell(1)
	ref.Tell(2)
	ref.Tell(3)
	ref.Tell(4)
	ref.Tell(5)
	ref.Tell(&Tell{mailbox: externalMailbox})

	count, err := externalMailbox.ReceiveTimeout(100 * time.Millisecond)
	if assert.NoError(t, err) {
		assert.Equal(t, 5, count)
	}
}
