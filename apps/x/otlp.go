package x

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type OTLPConfig struct {
	Endpoint string
	Labels   []attribute.KeyValue
	Metrics  Metrics
}
type metricsType int

const (
	Push metricsType = iota
	Pull
)

// TODO : have sensible defaults via an option interface
type Metrics struct {
	Type               metricsType
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

	resources := resource.Merge(resource.Default(),
		resource.NewWithAttributes(config.Labels...))

	ret := &instance{
		resources:    resources,
		promregistry: prom.NewRegistry(),
	}

	exporter, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(config.Endpoint),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	))

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
	if config.Metrics.Type == Push {
		metricController := controller.New(
			processor.New(
				simple.NewWithExactDistribution(), exporter,
			),
			controller.WithCollectPeriod(2*time.Second),
			controller.WithExporter(exporter),
			controller.WithResource(resources),
		)

		err = metricController.Start(ctx)

		if err != nil {
			return ret, fmt.Errorf("failed to start metric controller: %v", err)
		}

		global.SetMeterProvider(metricController.MeterProvider())
		ret.meterProvider = metricController.MeterProvider()

		ret.disposables = append(ret.disposables, func(ctx context.Context) error { return metricController.Stop(ctx) })
	}
	if config.Metrics.Type == Pull {

		exporter, err := prometheus.InstallNewPipeline(prometheus.Config{Registry: ret.PrometheusRegistry()})

		if err != nil {
			return ret, fmt.Errorf("failed to initialize prometheus exporter %v", err)
		}

		http.HandleFunc("/", exporter.ServeHTTP)

		go func() {
			_ = http.ListenAndServe(fmt.Sprintf(":%d", config.Metrics.Port), nil)
		}()
		global.SetMeterProvider(exporter.MeterProvider())
		ret.meterProvider = exporter.MeterProvider()
	}
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
