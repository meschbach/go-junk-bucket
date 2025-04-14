package actors

import "context"

// Runtime represents the world from teh point of view of an actor.  Actors must use this interface to work within the
// actor system.
type Runtime interface {
	Ingestor
	//Self results in the Pid of the executing actor
	Self() Pid
	//Log provides a structured mechanism for producing output.
	Log() Logger

	//Spawn a new actor delegating to actor for user messages with the given option set
	Spawn(actor MessageActor, opts ...any) Pid
	//SpawnPort creates a new linked port
	//TODO: Is this the correct place?
	SpawnPort() Port

	//SpawnMonitor starts a new actor dispatching to actor, returning Pid.  Scheduling location is up the runtime.
	//Deprecated: use Spawn(actor, MonitoringOpt{tell: who})
	SpawnMonitor(actor MessageActor) Pid
	//diff between mailbox and port: port is intended for external commms, mailbox is meant for rpc like comms
	SpawnMailbox() Port

	//Monitor2 allows external monitoring between two processes
	Monitor2(watching Pid, watcher Pid)
	//Unmonitor will remove the {watched,watcher} pairs.
	Unmonitor(watched Pid, watcher Pid)

	Terminate(target Pid)
	Exit(result any)

	Register(name string, who Pid)
	Unregister(name string)
	LookupPath(path string) Pid
	NamedRef(name string) string
	SelfNamedRef() string

	//Context is the context of the currently invoking tick
	Context() context.Context
}
