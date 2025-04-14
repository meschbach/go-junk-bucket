package actors

// Start indicates the actor is starting execution and should perform any in actor initialization
type Start struct{}

// PanicExit is a signal to indicate a monitored process has failed
type PanicExit struct {
	//Who is the actor which panicked
	Who Pid
	//Momento is the monitors momento
	Momento any
}

func NewPanicExit(who Pid, momento any) PanicExit {
	if who.Process == 0 {
		panic("pid is invalid")
	}
	return PanicExit{Who: who, Momento: momento}
}

// NormalExit indicates an actor exited with a given value
type NormalExit struct {
	//Who is the actor which exited
	Who Pid
	//ExitValue is the exiting value provided by the actor
	ExitValue any
	//Momento is state details
	Momento any
}
