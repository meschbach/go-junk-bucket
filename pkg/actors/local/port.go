package local

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"sync/atomic"
	"time"
)

const (
	portInit = iota
	portOpen
	portClosed
)

type port struct {
	self    actors.Pid
	mailbox chan any
	theater *system
	state   uint32
}

func newPort(self actors.Pid, theater *system) *port {
	return &port{
		self:    self,
		mailbox: make(chan any, 16),
		theater: theater,
		state:   portOpen,
	}
}

func (p *port) Pid() actors.Pid {
	return p.self
}

func (p *port) ReceiveChannel() <-chan any {
	return p.mailbox
}

func (p *port) ReceiveTimeout(wait time.Duration) (any, error) {
	ctx, finish := context.WithTimeout(context.Background(), wait)
	defer finish()

	select {
	case <-ctx.Done():
		return nil, &MessageTimeoutError{
			Waited: wait,
			For:    p.self,
		}
	case value := <-p.mailbox:
		return value, nil
	}
}

func (p *port) ReceiveWith(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, &MessageTimeoutError{
			Waited: 0,
			For:    p.self,
		}
	case value := <-p.mailbox:
		return value, nil
	}
}

//TODO: tracing -- is it feasible to do here?
func (p *port) told(from context.Context, m any) {
	if v := atomic.LoadUint32(&p.state); v == portOpen {
		p.mailbox <- m
	}
}

func (p *port) Tell(ctx context.Context, who actors.Pid, what any) {
	tracer := otel.Tracer(TracerName)
	base, done := context.WithCancel(ctx)
	defer done()
	portContext, span := tracer.Start(base, "port: "+reflect.TypeOf(what).String(), trace.WithSpanKind(trace.SpanKindProducer))
	span.SetAttributes(attribute.String("pid", p.self.String()))
	span.SetAttributes(attribute.String("telling", who.String()))
	defer span.End()
	p.theater.Tell(portContext, who, what)
}

func (p *port) Log() actors.Logger {
	return &consoleLogger{who: p.self}
}

func (p *port) Close(ctx context.Context) {
	if old := atomic.SwapUint32(&p.state, portClosed); old == portClosed {
		return
	}
	p.theater.removeTarget(p.self)
	close(p.mailbox)
}

func (p *port) Receive() any {
	return <-p.mailbox
}

type MessageTimeoutError struct {
	Waited time.Duration
	For    actors.Pid
}

func (m *MessageTimeoutError) Error() string {
	return fmt.Sprintf("Timed out after %s waiting on a message at %#v", m.Waited.String(), m.For)
}
