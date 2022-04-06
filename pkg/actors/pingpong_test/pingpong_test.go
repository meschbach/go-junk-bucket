package pingpong_test

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type stringDirector struct {
	apply  actors.Pid
	inform actors.Pid
}

func (s *stringDirector) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case string:
		r.Tell(s.apply, &appendString{
			to:   msg,
			next: s.inform,
		})
	default:
		panic("bad message type")
	}
}

type appender struct {
	suffix string
}

func (a *appender) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case *appendString:
		r.Tell(msg.next, msg.to+a.suffix)
	default:
		panic("bad message type")
	}
}

type appendString struct {
	to   string
	next actors.Pid
}

func TestPingPongActors(t *testing.T) {
	sys := local.NewSystem()
	ping := sys.SimpleSpawn(&appender{suffix: "ping"})
	pong := sys.SimpleSpawn(&appender{suffix: "pong"})
	out := sys.NewPort()
	director := sys.SimpleSpawn(&stringDirector{
		apply:  pong,
		inform: out.Pid(),
	})

	sys.Tell(ping, &appendString{
		to:   "a",
		next: director,
	})
	v, err := out.ReceiveTimeout(100 * time.Millisecond)
	if assert.NoError(t, err) {
		assert.Equal(t, "apingpong", v)
	}
}
