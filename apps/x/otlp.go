package x

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type OTLPConfig struct {
	Endpoint string
	Labels   []attribute.KeyValue
	Metrics  Metrics
}

// TODO : have sensible defaults via an option interface
type Metrics struct {
	Port               int
	IncludeHostMetrics bool
}

type instance struct {
	disposables   []func(context.Context) error
	resources     *resource.Resource
	meterProvider metric.MeterProvider
	promregistry  *prom.Registry
}

func (i *instance) Close(ctx context.Context) error {
	for d := range i.disposables {
		if err := i.disposables[d](ctx); err != nil {
			return err
		}
	}
	return nil
}

func (i *instance) Resources() *resource.Resource {
	return i.resources
}

func (i *instance) MeterProvider() metric.MeterProvider {
	return i.meterProvider
}
func (i *instance) PrometheusRegistry() *prom.Registry {
	return i.promregistry
}

func InitialiseOTLP(ctx context.Context, config OTLPConfig) (*instance, error) {

	resources := resource.NewWithAttributes(semconv.SchemaURL, config.Labels...)

	ret := &instance{
		resources:    resources,
		promregistry: prom.NewRegistry(),
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)

	if err != nil {
		return ret, fmt.Errorf("failed to create exporter: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resources),
		sdktrace.WithSpanProcessor(bsp),
	)
	propagators := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	ret.disposables = append(ret.disposables, func(ctx context.Context) error {
		return tracerProvider.Shutdown(ctx)
	})
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagators)

	if config.Metrics.IncludeHostMetrics {
		err = host.Start()
		if err != nil {
			return ret, fmt.Errorf("failed to start host instrumentation: %v", err)
		}
	}

	c := prometheus.Config{
		Registry: ret.promregistry,
	}
	metricController := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(c.DefaultHistogramBoundaries),
			),
			export.CumulativeExportKindSelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(resources),
	)
	promexporter, err := prometheus.New(c, metricController)
	if err != nil {
		log.Panicf("failed to initialize prometheus exporter %v", err)
	}
	global.SetMeterProvider(promexporter.MeterProvider())

	mux := http.NewServeMux()
	mux.HandleFunc("/", promexporter.ServeHTTP)
	go func() {
		if err = http.ListenAndServe(fmt.Sprintf(":%d", config.Metrics.Port), mux); err != nil {
			log.Panicf("failed to listen start prometheus service  %v", err)
		}
	}()

	global.SetMeterProvider(promexporter.MeterProvider())
	ret.meterProvider = promexporter.MeterProvider()

	ret.disposables = append(ret.disposables, func(ctx context.Context) error {
		return exporter.Shutdown(ctx)
	})

	return ret, nil

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

	bag := baggage.FromContext(ctx)

	if !ok {
		if v := bag.Member(correlationLabel).Value(); v != "" {
			c = v
		} else {
			c = uuid.New().String()
		}
	}

	// save to baggage and current trace
	cidLabel := attribute.String(correlationLabel, c)
	span.SetAttributes(cidLabel)

	// errors ignored as we are saving a UUID and the format and content
	// are known to be valid
	mem, _ := baggage.NewMember(correlationLabel, c) // nolint: errcheck
	bag, _ = bag.SetMember(mem)                      // nolint: errcheck

	ctx = baggage.ContextWithBaggage(ctx, bag)

	// create logger with trace and correlation id
	fields := map[string]interface{}{
		traceLabel:       span.SpanContext().TraceID(),
		correlationLabel: c,
	}
	// TODO : create a logger properly
	lgr := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).Level(zerolog.InfoLevel).With().Fields(fields).Timestamp().Logger()
	return lgr, ctx
}
