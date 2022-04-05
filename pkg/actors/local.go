package actors

import "sync"

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

type stage struct {
	mailbox chan any
}

func (l *LocalActorSystem) Spawn(name string, handler Actor) ActorRef {
	l.gracefulCompletionGate.Add(1)
	mailbox := make(chan any, 16)
	l.localActors[name] = &stage{
		mailbox: mailbox,
	}
	go func() {
		defer l.gracefulCompletionGate.Done()
		for m := range mailbox {
			handler.OnMessage(m)
		}
	}()
	return &localNamedRef{
		system: l,
		name:   name,
	}
}

func (l *LocalActorSystem) SpawnFn(name string, perform func(msg any)) ActorRef {
	return l.Spawn(name, &FnActor{perform: perform})
}
