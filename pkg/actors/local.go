package actors

import (
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/dispatcher"
	"sync"
)

type LocalActorSystem struct {
	gracefulCompletionGate sync.WaitGroup
	//todo: lock? actor?
	localActors map[string]*stage
}

func NewLocalActorSystem() *LocalActorSystem {
	return &LocalActorSystem{
		gracefulCompletionGate: sync.WaitGroup{},
		localActors:            make(map[string]*stage),
	}
}

//Shutdown gracefully shuts down all actors
func (l *LocalActorSystem) Shutdown() {
	for _, actor := range l.localActors {
		close(actor.mailbox)
	}
	l.gracefulCompletionGate.Wait()
}

//stage controls the integration of an actor into the larger system
type stage struct {
	mailbox chan any
	links   *dispatcher.Dispatcher[LinkEvent]
}

func (l *LocalActorSystem) Spawn(name string, handler Actor) ActorRef {
	l.gracefulCompletionGate.Add(1)
	mailbox := make(chan any, 16)
	actor := &stage{
		mailbox: mailbox,
		links:   dispatcher.NewDispatcher[LinkEvent](),
	}
	l.localActors[name] = actor

	ref := &localNamedRef{
		system: l,
		name:   name,
	}

	go func() {
		defer func() {
			defer l.gracefulCompletionGate.Done()
			problem := recover()
			if problem == nil {
				actor.links.Broadcast(LinkEvent{Actor: ref, Change: ActorDied})
			} else {
				actor.links.Broadcast(LinkEvent{Actor: ref, Change: ActorDone})
			}
		}()
		actor.links.Broadcast(LinkEvent{Actor: ref, Change: ActorRunning})
		for m := range mailbox {
			handler.OnMessage(m)
		}
	}()
	return ref
}

func (l *LocalActorSystem) SpawnFn(name string, perform func(msg any)) ActorRef {
	return l.Spawn(name, &FnActor{perform: perform})
}

type NoSuchActor struct {
	Name string
}

func (n *NoSuchActor) Error() string {
	return fmt.Sprintf("no such actor %q", n.Name)
}

func (l *LocalActorSystem) Monitor(rawRef ActorRef, mailbox *LocalMailbox) (func(), error) {
	ref := rawRef.(*localNamedRef)
	if ref.system != l {
		return nil, errors.New("not same actor system....need to figure that one out")
	}
	actor, ok := l.localActors[ref.name]
	if !ok {
		return nil, &NoSuchActor{Name: ref.name}
	}
	changes, done := actor.links.Listen()
	go func() {
		for {
			select {
			case <-mailbox.Done():
				done()
			case m := <-changes:
				mailbox.Tell(m)
			}
		}
	}()
	return done, nil
}

type ActorChange = uint8

const (
	ActorRunning ActorChange = iota
	ActorDied
	ActorDone
)

type LinkEvent struct {
	Actor  ActorRef
	Change ActorChange
}
