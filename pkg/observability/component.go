package observability

import (
	"context"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Component struct {
	otelAnchor *trace.TracerProvider
}

func (c *Component) ShutdownGracefully(ctx context.Context) error {
	if err := c.otelAnchor.ForceFlush(ctx); err != nil {
		return err
	}
	if err := c.otelAnchor.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
