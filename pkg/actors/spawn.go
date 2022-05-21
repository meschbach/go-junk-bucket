package actors

type MonitorOpt struct {
	Tell    Pid
	Momento any
}

type RegisterOpt struct {
	Name string
}
