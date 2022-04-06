package local

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"time"
)

type port struct {
	self    actors.Pid
	mailbox chan any
}

func newPort(self actors.Pid) *port {
	return &port{
		self:    self,
		mailbox: make(chan any, 16),
	}
}

func (p *port) Pid() actors.Pid {
	return p.self
}

func (p *port) ReceiveTimeout(wait time.Duration) (any, error) {
	ctx, finish := context.WithTimeout(context.Background(), wait)
	defer finish()

	select {
	case <-ctx.Done():
		return nil, &MessageTimeoutError{
			Waited: wait,
			For:    p.self,
		}
	case value := <-p.mailbox:
		return value, nil
	}
}

func (p *port) told(m any) {
	p.mailbox <- m
}

type MessageTimeoutError struct {
	Waited time.Duration
	For    actors.Pid
}

func (m *MessageTimeoutError) Error() string {
	return fmt.Sprintf("Timed out after %s waiting on a message at %#v", m.Waited.String(), m.For)
}
