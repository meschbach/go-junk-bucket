package supervisor

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type childState struct {
	pid actors.Pid
}

type supervisorStatus uint8

const (
	supervisorInit = iota
	supervisorAlive
	supervisorTerminatingForRestart
)

type actor struct {
	controller Behavior
	children   map[string]*childState
	status     supervisorStatus
	listeners  []actors.Pid
}

func newActor(controller Behavior) *actor {
	return &actor{
		controller: controller,
		children:   make(map[string]*childState),
		status:     supervisorInit,
		listeners:  make([]actors.Pid, 0),
	}
}

func (a *actor) OnMessage(r actors.Runtime, m any) {
	switch msg := m.(type) {
	case *actors.Start:
		a.start(r)
	case actors.Start:
		a.start(r)
	case actors.PanicExit:
		a.onPanic(r, msg)
	case actors.NormalExit:
		a.onExit(r, msg)
	case WatchState:
		a.listeners = append(a.listeners, msg.Observer)
	default:
		r.Log().Warn("unexpected message %#v", m)
	}
}

func (a *actor) start(r actors.Runtime) {
	self := r.Self()
	spec := a.controller.Init(r)
	a.status = supervisorAlive

	for _, c := range spec.Children {
		id := c.Id
		if _, has := a.children[id]; has {
			r.Log().Fatal("supervisor ID conflict: %s", id)
		} else {
			pid := r.Spawn(c.Start(), actors.MonitorOpt{Tell: self, Momento: c.Id})
			r.Register(id, pid)
			a.children[id] = &childState{pid: pid}
		}
	}
	//notify listeners
	for _, l := range a.listeners {
		r.Tell(l, StateReady{})
	}
}

func (a *actor) onPanic(r actors.Runtime, exit actors.PanicExit) {
	id := exit.Momento.(string)
	r.Log().Warn("actor %s (%s) panicked.", id, exit.Who)
	if _, has := a.children[id]; has {
		delete(a.children, id)
		r.Unregister(id)
		//restart all
		a.terminateForRestart(r)
	} else {
		r.Log().Warn("received panic for %s (pid: %s) but unknown", exit.Who, id)
	}
}

func (a *actor) terminateForRestart(r actors.Runtime) {
	a.status = supervisorTerminatingForRestart
	if len(a.children) == 0 {
		a.start(r)
	} else {
		for _, s := range a.children {
			r.Terminate(s.pid)
		}
	}
}

func (a *actor) onExit(r actors.Runtime, msg actors.NormalExit) {
	id := msg.Momento.(string)
	if _, has := a.children[id]; has {
		delete(a.children, id)
		r.Unregister(id)

		switch a.status {
		case supervisorAlive:
			a.terminateForRestart(r)
		case supervisorTerminatingForRestart:
			if len(a.children) == 0 {
				a.start(r)
			}
		default:
			r.Log().Fatal("unknown state for handling terminated actor: %d", a.status)
		}
	} else {
		r.Log().Warn("unknown id %q exited from %s", id, msg.Who)
	}
}
