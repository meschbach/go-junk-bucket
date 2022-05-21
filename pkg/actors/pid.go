package actors

import "fmt"

type Pid struct {
	Node    uint64
	Process uint64
}

func (p Pid) String() string {
	return fmt.Sprintf("<%d.%d>", p.Node, p.Process)
}

func (p *Pid) IsNil() bool {
	return p.Node == 0 && p.Process == 0
}
