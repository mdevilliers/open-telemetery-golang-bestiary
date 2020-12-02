package x

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TODO : urrgh get rid of package level function
func IntialiseTracing(name string, labels ...label.KeyValue) (func(), error) {
	f, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://0.0.0.0:14268/api/traces"), // NOTE this is the URL of the open-telemetary agent
		jaeger.WithProcess(jaeger.Process{
			ServiceName: name,
			Tags:        labels,
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)

	if err != nil {
		return func() {}, fmt.Errorf("failed to create exporter: %v", err)
	}

	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	return f, err

}
