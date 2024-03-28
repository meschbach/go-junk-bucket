package observability

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func newJaegerExporter(url string) (trace.SpanExporter, error) {
	// Create the Jaeger exporter
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
}

// newResource returns a resource describing this application.
func newResource(cfg Config) *resource.Resource {
	//TODO: properly resolve the version and environment
	r, problem := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceNameKey.String(cfg.ServiceName+"."+cfg.Environment),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", cfg.Environment),
		),
	)
	if problem != nil {
		panic(problem)
	}
	return r
}
