package registry

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type Registry struct {
	named map[string]actors.Pid
}

func NewRegistry() *Registry {
	return &Registry{named: make(map[string]actors.Pid)}
}

func (r *Registry) OnMessage(bif actors.Runtime, m any) {
	bif.Log().Info("registry <- %#v", m)
	switch msg := m.(type) {
	case *actors.Start:
	case Register:
		r.register(bif, msg)
	case Unregister:
		r.unregister(bif, msg)
	case Lookup:
		r.lookup(bif, msg)
	case actors.RpcCall[Queryable, LookupResult]:
		msg.Perform(bif, r)
	default:
		bif.Log().Warn("Unknown registry message: %#v", msg)
	}
}

func (r *Registry) register(bif actors.Runtime, msg Register) {
	name := msg.Name
	if _, contained := r.named[name]; contained {
		bif.Log().Warn("Name %q already registered, replacing", name)
	}
	r.named[name] = msg.Who
}

func (r *Registry) unregister(bif actors.Runtime, msg Unregister) {
	who, exists := r.named[msg.Name]
	if !exists {
		return
	}
	if who == msg.Who {
		delete(r.named, msg.Name)
	} else {
		bif.Log().Warn("Requested %q unregister but does not match Pid %s", msg.Name, msg.Who)
	}
}

func (r *Registry) lookup(bif actors.Runtime, msg Lookup) {
	who, exists := r.named[msg.Name]
	//TODO: client can cause a crash...perhaps there is a way to tell it to ignore problems?
	bif.Tell(msg.Tell, LookupResult{
		Found: exists,
		Name:  msg.Name,
		Who:   who,
	})
}

func (r *Registry) Lookup(runtime actors.Runtime, name string) LookupResult {
	who, exists := r.named[name]
	return LookupResult{
		Found: exists,
		Name:  name,
		Who:   who,
	}
}

type lookupProxy struct {
	registry actors.Pid
}

func (l *lookupProxy) OnMessage(bif actors.Runtime, m any) {
	switch m.(type) {
	case *actors.Start:
	case Lookup:
		bif.Tell(l.registry, m)
	case actors.RpcAction[Queryable, LookupResult]:
		bif.Tell(l.registry, m)
	default:
		bif.Log().Warn("LookupProxy received unexpected message: %#v", m)
	}
}

type SpawnedRegistry struct {
	Control actors.Pid
	Query   actors.Pid
}

func SpawnRegistry(bif actors.Runtime) SpawnedRegistry {
	//TODO: these should really be SpawnLink type things.
	registry := bif.SpawnMonitor(NewRegistry())
	proxy := bif.SpawnMonitor(&lookupProxy{registry: registry})
	return SpawnedRegistry{
		Control: registry,
		Query:   proxy,
	}
}
