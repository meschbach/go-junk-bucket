package local

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/dispatcher"
	"runtime/debug"
	"sync"
)

type stageState = uint8

const (
	stageStateInit = iota
	stageStateRunning
	stageStateStopping
	stageStateDone
)

type localGracefulShutdown struct {
}

//stage controls the integration of an actor into the larger system
type stage struct {
	mailbox chan any
	links   *dispatcher.Dispatcher[*actors.LinkEvent]
	name    string
	theater *ActorSystem
	state   stageState
}

func (s *stage) SpawnMonitor(name string, actor actors.Actor) (func(), actors.Mailbox) {
	actorName := s.name + "/" + name
	ref := s.theater.Spawn(actorName, actor)
	done, err := s.theater.Monitor(ref, s)
	if err != nil {
		panic(err)
	}
	return done, ref
}

func (s *stage) Exit() {
	switch s.state {
	case stageStateRunning:
		if s.state == stageStateRunning {
			s.doShutdown()
		}
	default:
		panic("not running")
	}
}

func (s *stage) doShutdown() {
	switch s.state {
	case stageStateRunning:
	case stageStateStopping:
	case stageStateDone:
		return
	default:
		panic(fmt.Sprintf("in state %d, not able to shutdown", s.state))
	}
	s.state = stageStateDone
	s.links.Broadcast(&actors.LinkEvent{Actor: &namedRef{
		system: s.theater,
		name:   s.name,
	}, Change: stageStateDone})
	close(s.mailbox)
}

func (s *stage) run(actor actors.Actor, selfRef actors.Mailbox, gate *sync.WaitGroup) {
	if s.state != stageStateInit {
		panic("running actor which has already started")
	}
	s.state = stageStateRunning
	fmt.Printf("Running %q\n", s.name)
	defer func() {
		s.state = stageStateStopping
		fmt.Printf("Finishing %q\n", s.name)
		defer func() {
			s.state = stageStateDone
			gate.Done()
		}()
		problem := recover()
		if selfRef == nil {
			panic("selfRef is nil")
		}
		if problem == nil {
			s.doShutdown()
		} else {
			if selfRef == nil {
				panic("self ref is nil")
			}
			fmt.Printf("Actor %q crashed because:\n", s.name)
			debug.PrintStack()
			s.links.Broadcast(&actors.LinkEvent{Actor: selfRef, Change: actors.ActorDied})
		}
	}()
	s.links.Broadcast(&actors.LinkEvent{Actor: selfRef, Change: actors.ActorRunning})
	actor.OnMessage(actors.SelfStarted{
		Stage: s,
		Self:  selfRef,
	})
	for m := range s.mailbox {
		switch m.(type) {
		case localGracefulShutdown:
			switch s.state {
			case stageStateRunning:
				s.Exit()
			}
		default:
			actor.OnMessage(m)
		}
	}
}

func (s *stage) Tell(m any) {
	fmt.Printf("Tell %q %#v\n", s.name, m)
	switch msg := m.(type) {
	case *actors.LinkEvent:
		if msg == nil {
			panic("nil ref for actor link event")
		}
		if msg.Actor == nil {
			panic("actor is nil")
		}
	}
	s.mailbox <- m
}

func (s *stage) Done() chan any {
	return s.mailbox
}

func (s *stage) String() string {
	return fmt.Sprintf("local stage for %q", s.name)
}

func (s *stage) gracefulShutdown() {
	switch s.state {
	case stageStateDone:
	default:
		s.Tell(localGracefulShutdown{})
	}
}
