package actors

type localNamedRef struct {
	system *LocalActorSystem
	name string
}

func (l *localNamedRef) Tell(msg any) {
	actor := l.system.localActors[l.name]
	actor.mailbox <- msg
}
