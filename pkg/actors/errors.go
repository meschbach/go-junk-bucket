package actors

import "fmt"

type NoSuchActor struct {
	Name string
}

func (n *NoSuchActor) Error() string {
	return fmt.Sprintf("no such actor %q", n.Name)
}
