package local

import "fmt"

type namedRef struct {
	system *ActorSystem
	name   string
}

func (l *namedRef) Tell(msg any) {
	actor := l.system.localActors[l.name]
	actor.Tell(msg)
}

func (l *namedRef) Done() chan any {
	actor := l.system.localActors[l.name]
	return actor.Done()
}

func (l *namedRef) String() string {
	return fmt.Sprintf("local ref for actor %q", l.name)
}
