package actors

import "time"

type Port interface {
	Pid() Pid
	ReceiveTimeout(wait time.Duration) (any, error)
}
