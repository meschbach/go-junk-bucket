package actors

type MessageActor interface {
	OnMessage(r Runtime, m any)
}

type Runtime interface {
	Tell(p Pid, m any)
}

type System interface {
	NewPort() Port
	SimpleSpawn(actor MessageActor) Pid
	Tell(p Pid, m any)
}
