package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/XSAM/otelsql"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/api"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/x"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

type Env struct {
	OTLPEndpoint string `envconfig:"OTLP_ENDPOINT" default:"0.0.0.0:4317"`
	DBHost       string `envconfig:"DB_HOST" default:"0.0.0.0"`
	DBName       string `envconfig:"DB_NAME" default:"otsql_db"`
	DBUserName   string `envconfig:"DB_USERNAME" default:"otsql_user"`
	DBPassword   string `envconfig:"DB_PASSWORD" default:"otsql_password"`
}

var config Env

func main() {
	// listens on GRPC :9777
	// querys a postgres database

	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("error initilising config : %v:", err)
	}

	// initialise OTLP with some shared code
	ctx := context.Background()

	otlp, err := x.InitialiseOTLP(ctx, x.OTLPConfig{
		Endpoint: config.OTLPEndpoint,
		Labels: []attribute.KeyValue{
			semconv.ServiceNameKey.String("service-one"),
			semconv.ServiceVersionKey.String("3.4"),
			semconv.ServiceNamespaceKey.String("demo")},
		Metrics: x.Metrics{
			Port: 2222,
		},
	})

	if err != nil {
		log.Fatalf("error initilising tracing : %v:", err)
	}
	defer otlp.Close(ctx)

	// create a db connection
	var dsn = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", config.DBUserName, config.DBPassword, config.DBHost, config.DBName)

	// Register an OTel driver
	driverName, err := otelsql.Register("postgres")
	if err != nil {
		log.Fatalf("failed to register DB driver : %v", err)
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Fatalf("failed to open DB connection: %v", err)
	}

	defer db.Close()

	lis, err := net.Listen("tcp", ":9777")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcMetrics := grpc_prometheus.NewServerMetrics()
	otlp.PrometheusRegistry().MustRegister(grpcMetrics)

	// wrap the GRPC server with the open telemetery and prometheus interceptors
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpcMetrics.UnaryServerInterceptor(), otelgrpc.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(grpcMetrics.StreamServerInterceptor(), otelgrpc.StreamServerInterceptor()),
	)

	api.RegisterHelloServiceServer(s, &server{db: db})
	log.Println("service started!")
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

	lgr, ctx := x.GetRequestContext(ctx)
	lgr.Info().Msg("SayHello")

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
