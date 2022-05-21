package supervisor

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type StartChild = func() actors.MessageActor

type ChildSpec struct {
	Id    string
	Start StartChild
}

type Spec struct {
	Children []ChildSpec
}

type Behavior interface {
	Init(bif actors.Runtime) Spec
}

func FromBehavior(b Behavior) actors.MessageActor {
	return newActor(b)
}
