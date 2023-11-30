package observability

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Component struct {
	otelAnchor *trace.TracerProvider
}

func (c *Component) ShutdownGracefully(ctx context.Context) error {
	if c == nil {
		return nil
	}
	if err := c.otelAnchor.Shutdown(ctx); err != nil {
		return errors.Join(errors.New("otel shutdown"), err)
	}
	return nil
}
