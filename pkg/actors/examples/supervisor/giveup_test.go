package supervisor

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/meschbach/go-junk-bucket/pkg/actors/supervisor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGivesUp(t *testing.T) {
	t.Parallel()

	ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
	defer done()

	sys := local.NewSystem()
	port := sys.NewPort()

	sup := sys.Spawn(ctx, supervisor.FromBehavior(&givesUpSupervisor{}))
	sys.Tell(ctx, sup, supervisor.WatchState{Observer: port.Pid()})

	pid := sys.Lookup(ctx, "/target")
	sys.Tell(ctx, pid, increment{})
	sys.Tell(ctx, pid, increment{})

	sys.Tell(ctx, pid, tell{who: port.Pid()})
	value, err := port.ReceiveWith(ctx)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(2), value)

	sys.Tell(ctx, pid, giveUp{})
	for {
		msg, err := port.ReceiveWith(ctx)
		if err != nil {
			panic(err)
		}
		t.Log(fmt.Sprintf("Received %#v", msg))
		switch msg.(type) {
		case supervisor.StateReady:
			pid := sys.Lookup(ctx, "/target")
			sys.Tell(ctx, pid, tell{who: port.Pid()})
			value, err = port.ReceiveWith(ctx)
			if err != nil {
				panic(err)
			}
			assert.Equal(t, uint(0), value)
			return
		}
	}
}
