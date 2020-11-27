package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/api"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/x"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/label"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	// listens on http :8777
	// calls svc-one via grpc

	// initialise tracing with some shared code
	flush, err := x.IntialiseTracing("client-api", label.String("version", "1.1"))
	if err != nil {
		log.Fatalf("error initilising tracing : %v:", err)
	}
	defer flush()

	// set up GRPC client wrapping it with the Open Telemetry handlers
	conn, err := grpc.Dial(":9777", grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	client := api.NewHelloServiceClient(conn)

	helloHandler := func(w http.ResponseWriter, req *http.Request) {

		md := metadata.Pairs(
			"timestamp", time.Now().Format(time.StampNano),
			"client-id", "web-api-client-us-east-1",
			"user-id", "some-test-user-id",
		)

		// NOTE : we pass on the original context
		ctx := metadata.NewOutgoingContext(req.Context(), md)
		response, err := client.SayHello(ctx, &api.HelloRequest{Greeting: "World"})
		if err != nil {
			log.Fatalf("error when calling SayHello: %s", err)
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
