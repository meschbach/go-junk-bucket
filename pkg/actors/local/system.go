package local

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"sync"
	"sync/atomic"
)

type messageTarget interface {
	told(m any)
}

type system struct {
	nextID    uint64
	actorLock sync.RWMutex
	actors    map[actors.Pid]messageTarget
}

func (s *system) nextPID() actors.Pid {
	id := atomic.AddUint64(&s.nextID, 1)
	return actors.Pid{
		Node:    0,
		Process: id,
	}
}

func (s *system) registerTarget(pid actors.Pid, target messageTarget) {
	s.actorLock.Lock()
	defer s.actorLock.Unlock()
	s.actors[pid] = target
}

func (s *system) NewPort() actors.Port {
	pid := s.nextPID()
	p := newPort(pid)
	s.registerTarget(pid, p)
	return p
}

func (s *system) Tell(p actors.Pid, m any) {
	if actor, ok := s.actors[p]; ok {
		actor.told(m)
	} else {
		panic(fmt.Sprintf("no such actor %v", p))
	}
}

func (s *system) SimpleSpawn(a actors.MessageActor) actors.Pid {
	pid := s.nextPID()
	r := &runtime{
		system:   s,
		self:     pid,
		mailbox:  make(chan any, 16),
		consumer: a,
	}
	s.registerTarget(pid, r)
	r.start()
	return pid
}

func NewSystem() actors.System {
	return &system{
		nextID: 0,
		actors: make(map[actors.Pid]messageTarget),
	}
}
