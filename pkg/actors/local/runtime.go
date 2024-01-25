package local

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
)

const TracerName = "git.meschbach.com/mee/junk/actors"

type runtimeState uint

const (
	runtimeInit = iota
	runtimeStarting
	runtimeRunning
	runtimeDone
)

type runtime struct {
	changes    sync.Mutex
	system     *system
	self       actors.Pid
	mailbox    chan tracedDecorator
	consumer   actors.MessageActor
	monitoring []startMonitoring
	state      runtimeState
	names      map[string]actors.Pid
	parent     *runtime
}

func (r *runtime) told(from context.Context, m any) {
	r.submit(from, &userMessage{m: m})
}

func (r *runtime) start() {
	r.changes.Lock()
	defer r.changes.Unlock()

	r.state = runtimeStarting
	go r.run()
}

func (r *runtime) done() {
	r.changes.Lock()
	defer r.changes.Unlock()

	if r.state == runtimeDone {
		return
	}
	r.state = runtimeDone
	r.system.removeTarget(r.self)
	close(r.mailbox)
	r.mailbox = nil
}

func (r *runtime) submit(from context.Context, action runtimeMessage) {
	r.changes.Lock()
	defer r.changes.Unlock()

	span := trace.SpanFromContext(from)
	//TODO: crud optimistic locking...race conditions can occur
	switch r.state {
	case runtimeDone:
		//todo: should really just log a warning with the invoking actor
		span.AddEvent("submit-to-done", trace.WithAttributes(attribute.Stringer("telling", r.self), attribute.String("action", fmt.Sprintf("%#v", action))))
	default:
		span.AddEvent("submit-signal", trace.WithAttributes(attribute.Stringer("telling", r.self), attribute.String("action", action.name())))
		//todo: tracing layer probably should be optional
		r.mailbox <- traceDecorator(from, action)
	}
}

func (r *runtime) run() {
	defer func() {
		r.done()
	}()

	r.startRunning()
	for m := range r.mailbox {
		if !r.isRunning() {
			break
		}
		r.tick(m)
	}
}

func (r *runtime) startRunning() {
	r.changes.Lock()
	defer r.changes.Unlock()

	r.state = runtimeRunning
}

func (r *runtime) isRunning() bool {
	r.changes.Lock()
	defer r.changes.Unlock()
	return r.state == runtimeRunning
}

var tracer = otel.Tracer(TracerName)

func (r *runtime) tick(signal tracedDecorator) {
	tickBase, tickBaseDone := context.WithCancel(context.Background())
	defer tickBaseDone()
	parentContext := signal.baseContext(tickBase)
	tickContext, span := tracer.Start(parentContext, signal.next.name(), trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	span.SetAttributes(attribute.Stringer("pid", r.self), attribute.String("name", signal.name()))

	defer func() {
		problem := recover()
		if problem != nil {
			if err, ok := problem.(error); ok {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				e := errors.New(fmt.Sprintf("%#v", problem))
				span.RecordError(e)
				span.SetStatus(codes.Error, e.Error())
			}

			//TODO: this operation might fault
			nameParts := r.namedParts()
			span.SetAttributes(attribute.StringSlice("name", nameParts))
			name := "/" + strings.Join(nameParts, "/")

			logger := r.system.loggingStrategy.buildLogger(tickContext, r.self)
			stackTrace := debug.Stack()
			logger.Error("actor panic: %s -- %#v\n%s", name, problem, stackTrace)

			//notify listeners
			for _, l := range r.monitoring {
				r.system.Tell(tickContext, l.listener, actors.NewPanicExit(r.self, l.what))
			}
			r.done()
		}
	}()

	signal.next.execute(tickContext, r)
}

func (r *runtime) onActorExit(tickContext context.Context, result any) {
	r.done()
	r.state = runtimeDone
	for _, l := range r.monitoring {
		r.system.Tell(tickContext, l.listener, actors.NormalExit{
			Who:       r.self,
			ExitValue: result,
			Momento:   l.what,
		})
	}
}

func (r *runtime) namedParts() []string {
	if r.parent == nil {
		return []string{}
	} else {
		return append(r.parent.namedParts(), r.parent.findNameFor(r.self))
	}
}

//findForName locates the name for p within this actor.  If no name may be found "<annoymous>" is returned.
func (r *runtime) findNameFor(p actors.Pid) string {
	for name, pid := range r.names {
		if pid == p {
			return name
		}
	}
	return "<anonymous>"
}

type runtimeMessage interface {
	execute(tickContext context.Context, r *runtime)
	name() string
}

type userMessage struct {
	m any
}

func (u *userMessage) execute(ctx context.Context, r *runtime) {
	c := &container{tickContext: ctx, r: r}
	span := trace.SpanFromContext(ctx)
	span.SetName("message: " + reflect.TypeOf(u.m).String())
	span.SetAttributes(attribute.String("pid", r.self.String()))
	r.consumer.OnMessage(c, u.m)
}

func (u *userMessage) name() string {
	return "message: " + reflect.TypeOf(u.m).String()
}

//https://devandchill.com/posts/2021/12/go-step-by-step-guide-for-implementing-tracing-on-a-microservices-architecture-2/2/
type tracedDecorator struct {
	carrier map[string]string
	next    runtimeMessage
}

func (t *tracedDecorator) Get(key string) string {
	v, ok := t.carrier[key]
	if !ok {
		return ""
	}
	return v
}

func (t *tracedDecorator) Set(key string, value string) {
	t.carrier[key] = value
}

func (t *tracedDecorator) Keys() []string {
	i := 0
	r := make([]string, len(t.carrier))

	for k, _ := range t.carrier {
		r[i] = k
		i++
	}

	return r
}

func (t *tracedDecorator) baseContext(fromParent context.Context) context.Context {
	return otel.GetTextMapPropagator().Extract(fromParent, t)
}

//TODO: really should become an envelope for the call
func (t *tracedDecorator) name() string {
	return t.next.name()
}

func traceDecorator(ctx context.Context, msg runtimeMessage) tracedDecorator {
	decorator := tracedDecorator{carrier: make(map[string]string), next: msg}
	otel.GetTextMapPropagator().Inject(ctx, &decorator)
	return decorator
}

type actorExitSignal struct {
	result any
}

func (a actorExitSignal) name() string {
	return "actor-exiting"
}

func (a actorExitSignal) execute(ctx context.Context, r *runtime) {
	r.onActorExit(ctx, a.result)
}
