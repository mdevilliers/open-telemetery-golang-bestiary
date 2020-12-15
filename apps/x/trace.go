package x

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
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

	propagators := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(propagators)

	return f, err

}

const (
	correlationIDHeader = "correlationIDHeader"
)

func CorrelationIDFromContext(ctx context.Context) (string, context.Context) {
	cid := ctx.Value(correlationIDHeader)
	span := trace.SpanFromContext(ctx)
	c, ok := cid.(string)

	if !ok {

		// look in baggage
		if span.IsRecording() {
			cid := baggage.Value(ctx, label.Key("x.correlation-id"))
			c = cid.AsString()
		}
		// if still empty - make one
		if c == "" {
			c = uuid.New().String()
		}

	}

	ctx = context.WithValue(ctx, correlationIDHeader, c)

	// save to baggage and current trace
	cidLabel := label.String("x.correlation-id", c)
	ctx = baggage.ContextWithValues(ctx, cidLabel)
	span.SetAttributes(cidLabel)
	return c, ctx
}
