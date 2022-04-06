package local

import (
	"context"
	"errors"
	"time"
)

type LocalMailbox struct {
	mailbox     chan any
	doneChannel chan interface{}
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

func (l *LocalMailbox) Done() chan interface{} {
	return l.doneChannel
}

func (l *LocalMailbox) Close() {
	close(l.doneChannel)
}

func (l *LocalActorSystem) ExternalMailbox() *LocalMailbox {
	return &LocalMailbox{mailbox: make(chan any, 16), doneChannel: make(chan interface{})}
}
