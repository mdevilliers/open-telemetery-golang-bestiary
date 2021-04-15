package x

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// TODO : urrgh get rid of package level function
func InitialiseTracing(ctx context.Context, endpoint, name string, labels ...attribute.KeyValue) (func(), error) {

	resources := resource.NewWithAttributes(
		semconv.ServiceNameKey.String(name),
	)
	exporter, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(endpoint),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	))

	if err != nil {
		return func() {}, fmt.Errorf("failed to create exporter: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.Merge(resources, resource.NewWithAttributes(labels...))),
		sdktrace.WithSpanProcessor(bsp),
	)
	propagators := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagators)

	return func() {
		tracerProvider.Shutdown(ctx)
		exporter.Shutdown(ctx)
	}, err

}

const (
	correlationIDHeader = "correlationIDHeader"

	correlationLabel = "correlationID"
	traceLabel       = "traceID"
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
			cid := baggage.Value(ctx, attribute.Key(correlationLabel))
			if cid.Type() != attribute.INVALID {
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
	cidLabel := attribute.String(correlationLabel, c)
	ctx = baggage.ContextWithValues(ctx, cidLabel)
	span.SetAttributes(cidLabel)

	// create logger with trace and correlation id
	fields := map[string]interface{}{
		traceLabel:       span.SpanContext().TraceID(),
		correlationLabel: c,
	}
	// TODO : create a logger properly
	lgr := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).Level(zerolog.InfoLevel).With().Fields(fields).Timestamp().Logger()
	return lgr, ctx

}
