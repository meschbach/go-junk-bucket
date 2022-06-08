package actors

import (
	"context"
)

//Ingestor is capable of dispatching messages to pids
type Ingestor interface {
	//Tell sends m to p.  If p is not alive or does not exist then the runtime will panic within this actor.
	Tell(p Pid, m any)
}

type MessageActor interface {
	OnMessage(r Runtime, m any)
}

//System is intended to represent an entire node
type System interface {
	Tell(ctx context.Context, p Pid, m any)
	//NewPort creates a local port on the system
	NewPort() Port
	//Spawn a new actor delegating to actor for user messages with the given option set
	Spawn(ctx context.Context, actor MessageActor, opts ...any) Pid
	Lookup(ctx context.Context, absolutePath string) Pid
}

type Logger interface {
	Error(format string, args ...any)
	Fatal(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
}
