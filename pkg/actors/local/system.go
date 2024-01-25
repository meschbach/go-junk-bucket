package local

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"sync"
	"sync/atomic"
)

type messageTarget interface {
	told(from context.Context, m any)
}

type system struct {
	nextID          uint64
	actorLock       sync.RWMutex
	actors          map[actors.Pid]messageTarget
	root            *runtime
	loggingStrategy LoggingStrategy
}

func (s *system) nextPID() actors.Pid {
	id := atomic.AddUint64(&s.nextID, 1)
	return actors.Pid{
		Node:    0,
		Process: id,
	}
}

func (s *system) registerTarget(pid actors.Pid, target messageTarget) {
	s.actorLock.Lock()
	defer s.actorLock.Unlock()
	s.actors[pid] = target
}

func (s *system) removeTarget(target actors.Pid) {
	s.actorLock.Lock()
	defer s.actorLock.Unlock()
	if _, has := s.actors[target]; has {
		delete(s.actors, target)
	}
}

func (s *system) NewPort() actors.Port {
	pid := s.nextPID()
	p := newPort(pid, s)
	s.registerTarget(pid, p)
	return p
}

//TODO: similar to another spot, merge?
func (s *system) Tell(ctx context.Context, p actors.Pid, m any) {
	actor := s.pid2target(p)
	if actor == nil {
		span := trace.SpanFromContext(ctx)
		span.AddEvent("missing-target", trace.WithAttributes(attribute.String("target", p.String())))
	} else {
		actor.told(ctx, m)
	}
}

func (s *system) Spawn(context context.Context, a actors.MessageActor, opts ...any) actors.Pid {
	var monitoring []actors.MonitorOpt
	var parent *runtime = nil
	var registerAs *string = nil
	for _, opt := range opts {
		switch o := opt.(type) {
		case actors.MonitorOpt:
			monitoring = append(monitoring, o)
		case *actors.MonitorOpt:
			monitoring = append(monitoring, *o)
		case parentOpt:
			parent = o.who
		case actors.RegisterOpt:
			registerAs = &o.Name
		default:
			panic(fmt.Sprintf("unknown option type %#v", opt))
		}
	}

	pid := s.nextPID()
	r := &runtime{
		changes:  sync.Mutex{},
		system:   s,
		self:     pid,
		mailbox:  make(chan tracedDecorator, 16),
		consumer: a,
		state:    runtimeInit,
		names:    make(map[string]actors.Pid),
		parent:   parent,
	}
	r.changes.Lock()
	if s.root == nil {
		s.root = r
	}
	if r.parent != nil && registerAs != nil {
		r.parent.names[*registerAs] = pid
	}
	r.changes.Unlock()

	for _, m := range monitoring {
		r.submit(context, &startMonitoring{listener: m.Tell, what: m.Momento})
	}
	r.told(context, &actors.Start{})
	s.registerTarget(pid, r)
	r.start()
	return pid
}

func NewSystem(opts ...SystemOpts) actors.System {
	out := &system{
		nextID:          0,
		actors:          make(map[actors.Pid]messageTarget),
		loggingStrategy: &CompositeLoggingStrategy{Loggers: []LoggingStrategy{&ConsoleLoggingStrategy{}, &CompositeLoggingStrategy{}}},
	}
	for _, opt := range opts {
		opt.customizeSystem(out)
	}
	return out
}

func (s *system) pid2target(pid actors.Pid) messageTarget {
	s.actorLock.RLock()
	defer s.actorLock.RUnlock()
	return s.actors[pid]
}

func (s *system) execute(from context.Context, targetPID actors.Pid, action runtimeMessage) {
	//TODO: other types of targets should be supported
	//TODO: if target has exited this will generate a panic
	// - can be triggered by attempting to grant to a nonexistent user
	// - implemented a type test but not really the best choice
	rawTarget := s.pid2target(targetPID)
	if target, ok := rawTarget.(*runtime); ok {
		target.submit(from, action)
	} else {
		span := trace.SpanFromContext(from)
		span.AddEvent("no-such-pid", trace.WithAttributes(attribute.Stringer("pid", targetPID)))
	}
}

func (s *system) Lookup(ctx context.Context, absoluteName string) actors.Pid {
	boundary, span := otel.Tracer(TracerName).Start(ctx, "system.Lookup(name)")
	defer span.End()

	component := s.root.self

	mailbox := s.NewPort()
	defer mailbox.Close(boundary)
	resolver := mailbox.Pid()
	span.SetAttributes(attribute.Stringer("resolver", resolver), attribute.String("absolute-name", absoluteName))
	parts := strings.Split(absoluteName, "/")[1:]
	for index, part := range parts {
		span.AddEvent("lookup component", trace.WithAttributes(attribute.Stringer("pid", component), attribute.String("name", part)))
		s.execute(boundary, component, &lookupNamedComponent{component: part, tell: resolver})
		result := mailbox.Receive()
		switch msg := result.(type) {
		case foundName:
			component = msg.who
			span.AddEvent("found name", trace.WithAttributes(attribute.Stringer("pid", msg.who)))
		case noSuchName:
			panic(fmt.Sprintf("no such component %q (%d) in path %#v", part, index, parts))
		default:
			panic(result)
		}
	}
	span.AddEvent("resolved", trace.WithAttributes(attribute.Stringer("pid", component)))
	return component
}
