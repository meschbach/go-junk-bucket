package pingpong

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type appender struct {
	suffix string
}

func (a *appender) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case *actors.Start:
	case *appendString:
		r.Tell(msg.next, msg.to+a.suffix)
	default:
		r.Log().Warn("Unexpected message %#v", m)
	}
}

type appendString struct {
	to   string
	next actors.Pid
}
