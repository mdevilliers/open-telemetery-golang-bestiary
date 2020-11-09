package x

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/trace"
)

// TODO : urrgh get rig of package level function
func IntialiseTracing(name string, labels ...label.KeyValue) error {
	exporter, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint("http://0.0.0.0:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: name,
			Tags:        labels,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %v", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithConfig(trace.Config{DefaultSampler: trace.AlwaysSample()}),
		trace.WithSyncer(exporter),
	)
	global.SetTracerProvider(tp)
	global.SetTextMapPropagator(otel.NewCompositeTextMapPropagator(propagators.TraceContext{}, propagators.Baggage{}))
	return nil
}
