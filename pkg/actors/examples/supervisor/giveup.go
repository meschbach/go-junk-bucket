package supervisor

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/actors/supervisor"
)

type increment struct{}
type giveUp struct{}
type tell struct{ who actors.Pid }

type givesUpActor struct {
	value uint
}

func (g *givesUpActor) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case *actors.Start:
		r.Log().Info("Starting")
	case increment:
		g.value++
	case tell:
		r.Log().Info("Telling %s", r.Self())
		r.Tell(msg.who, g.value)
	case giveUp:
		panic("giving up")
	}
}

type givesUpSupervisor struct {
}

func (g *givesUpSupervisor) Init(bif actors.Runtime) supervisor.Spec {
	return supervisor.Spec{Children: []supervisor.ChildSpec{
		{Id: "target", Start: func() actors.MessageActor {
			return &givesUpActor{}
		}},
	}}
}
