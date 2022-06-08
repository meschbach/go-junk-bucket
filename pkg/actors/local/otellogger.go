package local

import (
	"context"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type otelLogger struct {
	span trace.Span
}

func (o *otelLogger) Info(format string, args ...any) {
	if o.span.IsRecording() {
		msg := fmt.Sprintf(format, args...)
		o.span.AddEvent("info", trace.WithAttributes(attribute.String("message", msg)))
	}
}

func (o *otelLogger) Warn(format string, args ...any) {
	if o.span.IsRecording() {
		msg := fmt.Sprintf(format, args...)
		o.span.AddEvent("warn", trace.WithAttributes(attribute.String("message", msg)))
	}
}

func (o *otelLogger) Fatal(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	o.span.RecordError(err)
	o.span.SetStatus(codes.Error, "fatal")
	panic(err)
}

func (o *otelLogger) Error(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	o.span.RecordError(err)
	o.span.SetStatus(codes.Error, "error")
}

type OTELLoggingStrategy struct {
}

func (o *OTELLoggingStrategy) buildLogger(ctx context.Context, who actors.Pid) actors.Logger {
	span := trace.SpanFromContext(ctx)
	return &otelLogger{span: span}
}

func (o *OTELLoggingStrategy) customizeSystem(s *system) {
	s.loggingStrategy = o
}
