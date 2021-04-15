package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/api"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/x"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Env struct {
	JaegerEndpoint string `envconfig:"JAEGER_ENDPOINT" default:"http://0.0.0.0:14268/api/traces"`
	SvcOneHost     string `envconfig:"SVC_HOST" default:"0.0.0.0"`
}

var config Env

func main() {
	// listens on http :8777
	// calls svc-one via grpc
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("error initilising config : %v:", err)
	}

	// intialise tracing with some shared code
	flush, err := x.InitialiseTracing(config.JaegerEndpoint, "client-api", attribute.String("version", "1.1"))
	if err != nil {
		log.Fatalf("error initilising tracing : %v:", err)
	}
	defer flush()

	// initilise some metrics
	err = x.IntialiseMetrics()
	if err != nil {
		log.Fatalf("error initilising metrics : %v:", err)
	}

	// set up GRPC client wrapping it with the Open Telemetry handlers
	conn, err := grpc.Dial(fmt.Sprintf("%s:9777", config.SvcOneHost), grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	client := api.NewHelloServiceClient(conn)

	helloHandler := func(w http.ResponseWriter, req *http.Request) {

		lgr, ctx := x.GetRequestContext(req.Context())
		lgr.Info().Msg("SayHello")

		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("span.attribute.foo", "span-attribute-bar"))

		md := metadata.Pairs(
			"timestamp", time.Now().Format(time.StampNano),
			"client-id", "web-api-client-us-east-1",
			"user-id", "some-test-user-id",
		)

		// NOTE : we pass on the original context
		ctx = metadata.NewOutgoingContext(ctx, md)
		response, err := client.SayHello(ctx, &api.HelloRequest{Greeting: "World"})
		if err != nil {
			lgr.Fatal().Err(err).Msg("error when calling SayHello")
		}
		_, _ = io.WriteString(w, fmt.Sprintf("%s\n", response))
	}

	// wrap http handler with generic tracer
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello")

	http.Handle("/hello", otelHandler)
	if err = http.ListenAndServe(":8777", nil); err != nil {
		panic(err)
	}
}
