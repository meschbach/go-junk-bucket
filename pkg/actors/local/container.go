package local

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

type container struct {
	r           *runtime
	tickContext context.Context
}

func (c *container) Self() actors.Pid {
	return c.r.self
}

func (c *container) Tell(p actors.Pid, m any) {
	c.r.system.Tell(c.tickContext, p, m)
}

func (c *container) SpawnPort() actors.Port {
	//TODO: if the runtime dies we should close the port
	p := c.r.system.NewPort()
	return p
}

func (c *container) Spawn(actor actors.MessageActor, opts ...any) actors.Pid {
	return c.r.system.Spawn(c.tickContext, actor, append(opts, parentOpt{who: c.r})...)
}

func (c *container) SpawnMonitor(actor actors.MessageActor) actors.Pid {
	return c.Spawn(actor, actors.MonitorOpt{
		Tell:    c.r.self,
		Momento: nil,
	})
}

func (c *container) Log() actors.Logger {
	return c.r.system.loggingStrategy.buildLogger(c.tickContext, c.r.self)
}

func (c *container) SpawnMailbox() actors.Port {
	return c.SpawnPort()
}

func (c *container) Monitor2(watched actors.Pid, watcher actors.Pid) {
	c.r.system.execute(c.tickContext, watched, &startMonitoring{listener: watcher})
}

func (c *container) Unmonitor(watched actors.Pid, watcher actors.Pid) {
	c.r.system.execute(c.tickContext, watched, &stopMonitoring{listener: watcher})
}

func (c *container) Terminate(who actors.Pid) {
	c.r.system.execute(c.tickContext, who, &terminateSignal{})
}

func (c *container) Exit(result any) {
	c.r.submit(c.tickContext, actorExitSignal{result: result})
}

func (c *container) Register(name string, who actors.Pid) {
	c.r.names[name] = who
}

func (c *container) Context() context.Context {
	return c.tickContext
}

func (c *container) Unregister(name string) {
	delete(c.r.names, name)
}

func (c *container) LookupPath(path string) actors.Pid {
	var component actors.Pid
	if strings.HasPrefix(path, "/") {
		path = path[1:]
		n := c.r
		for n.parent != nil {
			n = n.parent
		}
		component = n.self
	} else {
		component = c.r.self
	}

	mailbox := c.SpawnMailbox()
	resolver := mailbox.Pid()
	parts := strings.Split(path, "/")
	for index, part := range parts {
		c.r.system.execute(c.tickContext, component, &lookupNamedComponent{component: part, tell: resolver})
		result := mailbox.Receive()
		switch msg := result.(type) {
		case foundName:
			component = msg.who
		case noSuchName:
			c.Log().Fatal("no such component %q (%d) in path %#v", part, index, parts)
		default:
			panic(result)
		}
	}
	return component
}

type lookupNamedComponent struct {
	component string
	tell      actors.Pid
}

type foundName struct {
	who actors.Pid
}

type noSuchName struct {
}

func (l *lookupNamedComponent) execute(ctx context.Context, r *runtime) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("name", l.component), attribute.Stringer("tell", l.tell))
	if pid, has := r.names[l.component]; has {
		r.system.Tell(ctx, l.tell, foundName{who: pid})
	} else {
		r.system.Tell(ctx, l.tell, noSuchName{})
	}
}

func (l *lookupNamedComponent) name() string {
	return "lookup name"
}

func (c *container) NamedRef(name string) string {
	parts := c.r.namedParts()
	path := append(parts, name)
	pathName := "/" + strings.Join(path, "/")
	return pathName
}

func (c *container) SelfNamedRef() string {
	path := c.r.namedParts()
	pathName := "/" + strings.Join(path, "/")
	return pathName
}
