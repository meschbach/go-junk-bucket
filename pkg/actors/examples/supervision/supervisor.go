package supervision

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type supervisor struct {
	output actors.Mailbox
	count  uint
}

func (e *supervisor) OnMessage(msg any) {
	fmt.Printf("Supervisor: Received message: %v\n", msg)
	if msg == nil {
		panic("nil provided to onMessage")
	}
	switch m := msg.(type) {
	case actors.SelfStarted:
		fmt.Println("Starting supervisor")
		_, ref := m.Stage.SpawnMonitor("child", &exampleChild{})
		ref.Tell("every")
		ref.Tell("time")
		ref.Tell(get{inform: m.Self})
		ref.Tell(poison{})
	case *actors.LinkEvent:
		switch m.Change {
		case actors.ActorDied:
			e.output.Tell(e.count)
		}
	case uint:
		e.count = m
	default:
		panic("unexpected input")
	}
}
