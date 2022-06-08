package local

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type LoggingStrategy interface {
	buildLogger(ctx context.Context, who actors.Pid) actors.Logger
}

type ConsoleLoggingStrategy struct{}

func (c *ConsoleLoggingStrategy) buildLogger(ctx context.Context, who actors.Pid) actors.Logger {
	return &consoleLogger{who: who}
}

func (c *ConsoleLoggingStrategy) customizeSystem(s *system) {
	s.loggingStrategy = c
}
