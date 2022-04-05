package actors

type ActorRef interface {
	Tell(msg any)
}

type Actor interface {
	OnMessage(m any)
}
