package local

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type runtime struct {
	system   actors.System
	self     actors.Pid
	mailbox  chan any
	consumer actors.MessageActor
}

func (r *runtime) told(m any) {
	r.mailbox <- m
}

func (r *runtime) start() {
	go r.run()
}

func (r *runtime) run() {
	for m := range r.mailbox {
		r.consumer.OnMessage(r.system, m)
	}
}
