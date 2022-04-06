package local

import (
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/dispatcher"
	"sync"
)

type ActorSystem struct {
	gracefulCompletionGate sync.WaitGroup
	//todo: lock? actor?
	localActors map[string]*stage
}

func NewLocalActorSystem() *ActorSystem {
	return &ActorSystem{
		gracefulCompletionGate: sync.WaitGroup{},
		localActors:            make(map[string]*stage),
	}
}

//Shutdown gracefully shuts down all actors
func (l *ActorSystem) Shutdown() {
	for _, actor := range l.localActors {
		actor.gracefulShutdown()
	}
	l.gracefulCompletionGate.Wait()
}

func (l *ActorSystem) Spawn(name string, handler actors.Actor) actors.Mailbox {
	fmt.Printf("Spawning %q\n", name)
	l.gracefulCompletionGate.Add(1)
	mailbox := make(chan any, 16)
	actor := &stage{
		mailbox: mailbox,
		links:   dispatcher.NewDispatcher[*actors.LinkEvent](),
		name:    name,
		theater: l,
		state:   stageStateInit,
	}

	l.localActors[name] = actor

	ref := &namedRef{
		system: l,
		name:   name,
	}

	go actor.run(handler, ref, &l.gracefulCompletionGate)
	return ref
}

func (l *ActorSystem) SpawnFn(name string, perform func(msg any)) actors.Mailbox {
	return l.Spawn(name, actors.NewFnActor(perform))
}

func (l *ActorSystem) Monitor(rawRef actors.Mailbox, mailbox actors.Mailbox) (func(), error) {
	ref := rawRef.(*namedRef)
	if ref.system != l {
		return nil, errors.New("not same actor system....need to figure that one out")
	}
	actor, ok := l.localActors[ref.name]
	if !ok {
		return nil, &actors.NoSuchActor{Name: ref.name}
	}
	changes, done := actor.links.Listen()
	go func() {
		for {
			select {
			case <-mailbox.Done():
				done()
			case m := <-changes:
				fmt.Printf("Monitor relay for %v -- event %v\n ", mailbox, m)
				mailbox.Tell(m)
			}
		}
	}()
	return done, nil
}

func (l *ActorSystem) ExternalMailbox() *LocalMailbox {
	return &LocalMailbox{mailbox: make(chan any, 16), doneChannel: make(chan interface{})}
}
