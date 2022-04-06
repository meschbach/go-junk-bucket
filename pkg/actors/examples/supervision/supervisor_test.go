package actors

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type exampleSupervisor struct {
	output Mailbox
	count  uint
}

func (e *exampleSupervisor) OnMessage(msg any) {
	switch m := msg.(type) {
	case SelfStarted:
		_, ref := m.Stage.SpawnMonitor("child", &exampleChild{})
		ref.Tell(nil)
		ref.Tell(nil)
		ref.Tell(exampleSupervisorGet{inform: m.Self})
		ref.Tell(exampleSupervisorPanic{})
	case LinkEvent:
		switch m.Change {
		case ActorDied:
			e.output.Tell(e.count)
		}
	case uint:
		e.count = m
	}
}

type exampleSupervisorGet struct {
	inform Mailbox
}

type exampleSupervisorPanic struct {
}

type exampleChild struct {
	callCount uint
}

func (e *exampleChild) OnMessage(msg any) {
	fmt.Println("Calling child")
	switch m := msg.(type) {
	case exampleSupervisorGet:
		m.inform.Tell(e.callCount)
	case exampleSupervisorPanic:
		panic(m)
	default:
		e.callCount++
	}
}

func TestSupervisionTree(t *testing.T) {
	t.Parallel()

	sys := NewLocalActorSystem()
	defer sys.Shutdown()

	mailbox := sys.ExternalMailbox()
	defer mailbox.Close()
	sys.Spawn("supervisor", &exampleSupervisor{output: mailbox})
	value, err := mailbox.ReceiveTimeout(100 * time.Millisecond)
	if assert.NoError(t, err) {
		assert.Equal(t, 2, value)
	}
}
