package local

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type CompositeLogger struct {
	Loggers []actors.Logger
}

func (c *CompositeLogger) Info(format string, args ...any) {
	for _, l := range c.Loggers {
		l.Info(format, args...)
	}
}

func (c *CompositeLogger) Warn(format string, args ...any) {
	for _, l := range c.Loggers {
		l.Warn(format, args...)
	}
}

func (c *CompositeLogger) Fatal(format string, args ...any) {
	lastIndex := len(c.Loggers)
	dispatch := func(l actors.Logger) {
		defer func() {
			recover()
		}()
		l.Fatal(format, args...)
	}
	for _, l := range c.Loggers[0 : lastIndex-1] {
		go dispatch(l)
	}

	l := c.Loggers[lastIndex]
	l.Fatal(format, args...)
}

func (c *CompositeLogger) Error(format string, args ...any) {
	for _, l := range c.Loggers {
		l.Error(format, args...)
	}
}

type CompositeLoggingStrategy struct {
	Loggers []LoggingStrategy
}

func (c *CompositeLoggingStrategy) buildLogger(ctx context.Context, who actors.Pid) actors.Logger {
	var loggers []actors.Logger
	for _, strategy := range c.Loggers {
		loggers = append(loggers, strategy.buildLogger(ctx, who))
	}
	return &CompositeLogger{Loggers: loggers}
}

func (c *CompositeLoggingStrategy) customizeSystem(s *system) {
	s.loggingStrategy = c
}
