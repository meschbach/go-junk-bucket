package calculator_test

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCalculatorAdd(t *testing.T) {
	t.Parallel()

	sys := local.NewLocalActorSystem()
	response := sys.ExternalMailbox()
	ref := sys.Spawn("calculator", &calculator{value: 10})
	ref.Tell(&add{amount: 5})
	ref.Tell(&get{tell: response})

	value, err := response.ReceiveTimeout(100 * time.Millisecond)
	if assert.NoError(t, err) {
		assert.Equal(t, 15, value)
	}
	sys.Shutdown()
}

type calculator struct {
	value int
}

func (c *calculator) OnMessage(m any) {
	switch msg := m.(type) {
	case operation:
		msg.perform(c)
	case actors.SelfStarted:
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
	tell *local.LocalMailbox
}

func (g *get) perform(c *calculator) {
	g.tell.Tell(c.value)
}
