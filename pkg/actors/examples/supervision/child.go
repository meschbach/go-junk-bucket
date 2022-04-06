package supervision

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type get struct {
	inform actors.Mailbox
}

type poison struct {
}

type exampleChild struct {
	callCount uint
	stage     actors.Stage
	self      actors.Mailbox
}

func (e *exampleChild) OnMessage(msg any) {
	fmt.Printf("Child: %#v\n", msg)
	switch m := msg.(type) {
	case actors.SelfStarted:
		e.stage = m.Stage
		e.self = m.Self
	case get:
		fmt.Println("get invoked")
		m.inform.Tell(e.callCount)
	case poison:
		e.stage.Exit()
	default:
		e.callCount++
	}
}
