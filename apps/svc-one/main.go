package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/api"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/x"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/label"
	"google.golang.org/grpc"
)

func main() {
	// listens on GRPC :9777

	// initialise tracing with some shared code
	x.IntialiseTracing("service-one", label.String("version", "3.4"))

	lis, err := net.Listen("tcp", ":9777")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// wrap the GRPC server with the open telemetery handlers
	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	api.RegisterHelloServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

type server struct {
	api.HelloServiceServer
}

// SayHello implements api.HelloServiceServer
func (s *server) SayHello(ctx context.Context, in *api.HelloRequest) (*api.HelloResponse, error) {
	time.Sleep(50 * time.Millisecond)
	return &api.HelloResponse{Reply: "Hello " + in.Greeting}, nil
}
