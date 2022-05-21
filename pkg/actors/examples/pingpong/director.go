package pingpong

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type stringDirector struct {
	apply  actors.Pid
	inform actors.Pid
}

func (s *stringDirector) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case *actors.Start:
	case string:
		r.Tell(s.apply, &appendString{
			to:   msg,
			next: s.inform,
		})
	default:
		r.Log().Warn("Unexpected message %#v", m)
	}
}
