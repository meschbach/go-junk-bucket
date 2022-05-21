package registry

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type NamedRef struct {
	name  string
	query actors.Pid
}

func (n NamedRef) Resolve(bif actors.Runtime) actors.Pid {
	c := NewQueryClient(bif, n.query)
	return c.Find(n.name)
}
