package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/j2gg0s/otsql"
	"github.com/kelseyhightower/envconfig"
	pq "github.com/lib/pq"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/api"
	"github.com/mdevilliers/open-telemetery-golang-bestiary/apps/x"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/label"
	"google.golang.org/grpc"
)

type Env struct {
	JaegerEndpoint string `envconfig:"JAEGER_ENDPOINT" default:"http://0.0.0.0:14268/api/traces"`
	DBHost         string `envconfig:"DB_HOST" default:"0.0.0.0"`
	DBName         string `envconfig:"DB_NAME" default:"otsql_db"`
	DBUserName     string `envconfig:"DB_USERNAME" default:"otsql_user"`
	DBPassword     string `envconfig:"DB_PASSWORD" default:"otsql_password"`
}

var config Env

func main() {
	// listens on GRPC :9777
	// querys a postgres database

	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("error initilising config : %v:", err)
	}

	// initialise tracing with some shared code
	flush, err := x.InitialiseTracing(config.JaegerEndpoint, "service-one", label.String("version", "3.4"))
	if err != nil {
		log.Fatalf("error initilising tracing : %v:", err)
	}
	defer flush()

	// create a db connection
	var dsn = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", config.DBUserName, config.DBPassword, config.DBHost, config.DBName)

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
