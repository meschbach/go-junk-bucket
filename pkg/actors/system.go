package actors

type Mailbox interface {
	Tell(msg any)
	Done() chan any
}

type Actor interface {
	OnMessage(m any)
}

//Stage is how an actor interacts with it's controlling environment
type Stage interface {
	SpawnMonitor(name string, handler Actor) (func(), Mailbox)
	Exit()
}

//SelfStarted is sent to an actor on start to provide links into the execution environment
type SelfStarted struct {
	Stage Stage
	Self  Mailbox
}

type ActorChange = uint8

const (
	ActorNew ActorChange = iota
	ActorRunning
	ActorDied
	ActorDone
)

type LinkEvent struct {
	Actor  Mailbox
	Change ActorChange
}
