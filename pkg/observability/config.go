package observability

import (
	"github.com/meschbach/go-junk-bucket/pkg"
)

//Config is used to describe the setup of the observability libraries
type Config struct {
	//Exporter is the name of the OTEL exporter to utilize.  Currently supported values are `none` and `jaeger`.
	Exporter string `json:"exporter"`
	//ServiceName is the name to utilize in report tracing
	ServiceName string `json:"service-name"`
}

//DefaultConfig pulls values from the environment for the service or uses sensible defaults.
func DefaultConfig(serviceName string) Config {
	return Config{
		Exporter:    pkg.EnvOrDefault("OTEL_EXPORTER", "none"),
		ServiceName: pkg.EnvOrDefault("OTEL_SERVICE_NAME", serviceName),
	}
}
