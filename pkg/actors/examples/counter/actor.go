package counter

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
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
