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

//Behavior describes the children to be supervised and the resulting behaviors to exhibit in reaction to their state
type Behavior interface {
	//Init is responsible for building the child specification for the given environment
	Init(env actors.Runtime) Spec
}

//FromBehavior creates an actor to manage the described behaviors.
func FromBehavior(b Behavior) actors.MessageActor {
	return newActor(b)
}
