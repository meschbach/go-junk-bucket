package routing

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type router[M any, K any] struct {
	//TOOD: generic map
	routees map[any]actors.Pid
	bridge  Bridge[M, K]
}

func NewRouter[M any, K any](bridge Bridge[M, K]) actors.MessageActor {
	return &router[M, K]{
		routees: make(map[any]actors.Pid),
		bridge:  bridge,
	}
}

func (r *router[M, K]) OnMessage(bif actors.Runtime, m any) {
	switch msg := m.(type) {
	case M:
		r.route(bif, msg)
	}
}

func (r *router[M, K]) route(bif actors.Runtime, m M) {
	key := r.bridge.Extract(m)
	if target, ok := r.routees[key]; ok {
		bif.Tell(target, m)
	} else {
		spawned := r.bridge.SpawnRoutee(bif, key, m)
		r.routees[key] = spawned
		bif.Tell(spawned, m)
	}
}
