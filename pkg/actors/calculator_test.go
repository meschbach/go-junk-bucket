package actors

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type calculator struct {
	value int
}

func (c *calculator) OnMessage(m any) {
	switch msg := m.(type) {
	case operation:
		msg.perform(c)
	default:
		panic(m)
	}
}

type operation interface {
	perform(c *calculator)
}

type add struct {
	amount int
}

func (a *add) perform(c *calculator) {
	c.value += a.amount
}

type get struct {
	tell *LocalMailbox
}

func (g *get) perform(c *calculator) {
	g.tell.Tell(c.value)
}

func TestCalculatorAdd(t *testing.T) {
	t.Parallel()

	sys := NewLocalActorSystem()
	local := sys.ExternalMailbox()
	ref := sys.Spawn("calculator", &calculator{value: 10})
	ref.Tell(&add{amount: 5})
	ref.Tell(&get{tell: local})

	value, err := local.ReceiveTimeout(100 * time.Millisecond)
	if assert.NoError(t, err) {
		assert.Equal(t, 15, value)
	}
	sys.Shutdown()
}
