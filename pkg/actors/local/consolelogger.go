package local

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
)

type consoleLogger struct {
	who actors.Pid
}

func (c *consoleLogger) write(level string, format string, args []any) {
	newArgs := append([]any{c.who.String(), level}, args...)
	newFormat := "%s %s: " + format + "\n"
	fmt.Printf(newFormat, newArgs...)
}

func (c *consoleLogger) Info(format string, args ...any) {
	c.write("info", format, args)
}

func (c *consoleLogger) Warn(fmt string, args ...any) {
	c.write("warn", fmt, args)
}

func (c *consoleLogger) Fatal(format string, args ...any) {
	problem := fmt.Sprintf(format, args...)
	panic(problem)
}

func (c *consoleLogger) Error(format string, args ...any) {
	c.write("error", format, args)
}
