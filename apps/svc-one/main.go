package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	"github.com/j2gg0s/otsql"
	pq "github.com/lib/pq"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/api"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/x"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func main() {
	// listens on GRPC :9777
	// querys a postgres database

	// initialise tracing with some shared code
	flush, err := x.IntialiseTracing("service-one", label.String("version", "3.4"))
	if err != nil {
		log.Fatalf("error initilising tracing : %v:", err)
	}
	defer flush()

	// create a db connection
	var dsn = "postgres://otsql_user:otsql_password@localhost:5432/otsql_db?sslmode=disable"

	// create and wrap a DB connection
	connector, err := pq.NewConnector(dsn)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	db := sql.OpenDB(
		otsql.WrapConnector(connector, otsql.WithQuery(true)))
	defer db.Close()

	lis, err := net.Listen("tcp", ":9777")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// wrap the GRPC server with the open telemetery handlers
	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	api.RegisterHelloServiceServer(s, &server{db: db})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

type server struct {
	api.HelloServiceServer
	db *sql.DB
}

// SayHello implements api.HelloServiceServer
func (s *server) SayHello(ctx context.Context, in *api.HelloRequest) (*api.HelloResponse, error) {
	time.Sleep(25 * time.Millisecond)

	cid, ctx := x.CorrelationIDFromContext(ctx)
	log.Println("correlation id", cid)

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		log.Print("current trace id :", span.SpanContext().TraceID)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT NOW()`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var currentTime time.Time
	for rows.Next() {
		err = rows.Scan(&currentTime)
		if err != nil {
			return nil, err
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	time.Sleep(25 * time.Millisecond)

	return &api.HelloResponse{Reply: "Hello " + in.Greeting + "at" + currentTime.String()}, nil
}
