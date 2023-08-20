package streams

import "go.opentelemetry.io/otel"

var tracing = otel.Tracer("github.com/meschbach/go-junk-bucket/pkg/streams")
