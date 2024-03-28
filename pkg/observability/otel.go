package observability

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

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
