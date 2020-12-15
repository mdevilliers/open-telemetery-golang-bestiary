package x

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
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

	correlationLabel = "correlation-id"
	traceLabel       = "trace-id"
)

// GetRequestContext returns a logger and context populated with the current
// correlation and trace ids.
func GetRequestContext(ctx context.Context) (zerolog.Logger, context.Context) {

	cid := ctx.Value(correlationIDHeader)
	span := trace.SpanFromContext(ctx)
	c, ok := cid.(string)

	if !ok {
		// look in baggage
		if span.IsRecording() {
			cid := baggage.Value(ctx, label.Key(correlationLabel))
			if cid.Type() != label.INVALID {
				c = cid.AsString()
			}
		}
		// if still empty - make one
		if c == "" {
			c = uuid.New().String()
		}

	}

	ctx = context.WithValue(ctx, correlationIDHeader, c)

	// save to baggage and current trace
	cidLabel := label.String(correlationLabel, c)
	ctx = baggage.ContextWithValues(ctx, cidLabel)
	span.SetAttributes(cidLabel)

	// create logger with trace and correlation id
	fields := map[string]interface{}{
		traceLabel:       span.SpanContext().TraceID,
		correlationLabel: c,
	}
	// TODO : create a logger properly
	lgr := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Fields(fields).Timestamp().Logger()

	return lgr, ctx
}
