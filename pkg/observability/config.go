package observability

import (
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Config is used to describe the setup of the observability libraries
type Config struct {
	//Exporter is the name of the OTEL exporter to utilize.  Currently supported values are `none` and `jaeger`.
	Exporter string `json:"exporter"`
	//ServiceName is the name to utilize in report tracing
	ServiceName string `json:"service-name"`
	//Silent will not generate any output regarding the runtime configuration of the otel system
	Silent bool `json:"silent"`
}

// DefaultConfig pulls values from the environment for the service or uses sensible defaults.
func DefaultConfig(serviceName string) Config {
	return Config{
		Exporter:    pkg.EnvOrDefault("OTEL_EXPORTER", "none"),
		ServiceName: pkg.EnvOrDefault("OTEL_SERVICE_NAME", serviceName),
	}
}

func (c Config) Start() (*Component, error) {
	//not sure why this manually needs to be configured; perhaps need to look into others
	otel.SetTextMapPropagator(propagation.TraceContext{})

	if !c.Silent {
		fmt.Printf("Tracing %#v\n", c)
	}
	var exp trace.SpanExporter
	var err error
	switch c.Exporter {
	case "jaeger":
		jaegerEndpoint := pkg.EnvOrDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
		if !c.Silent {
			fmt.Printf("Using %q for Jaeger endpoint\n", jaegerEndpoint)
		}
		exp, err = newJaegerExporter(jaegerEndpoint)
	case "none":
		if !c.Silent {
			fmt.Println("No exporter configured.  Not recording spans.")
		}
	default:
		err = &UnknownExportError{Exporter: c.Exporter}
	}
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource(c)),
	)
	otel.SetTracerProvider(tp)
	return &Component{otelAnchor: tp}, nil
}

type UnknownExportError struct {
	Exporter string
}

func (u *UnknownExportError) Error() string {
	return fmt.Sprintf("Unknown export: %q", u.Exporter)
}
