package actors

import (
	"context"
	"errors"
	"time"
)

type LocalMailbox struct {
	mailbox chan any
}

func (l *LocalMailbox) Tell(m any) {
	l.mailbox <- m
}

func (l *LocalMailbox) ReceiveTimeout(amount time.Duration) (any, error) {
	timeout, done := context.WithTimeout(context.Background(), amount)
	defer done()

	select {
	case response := <-l.mailbox:
		return response, nil
	case <-timeout.Done():
		return nil, errors.New("timeout")
	}
}

func (l *LocalMailbox) Receive() any {
	return <-l.mailbox
}

func (l *LocalActorSystem) ExternalMailbox() *LocalMailbox {
	return &LocalMailbox{mailbox: make(chan any, 16)}
}
