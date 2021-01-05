module github.com/mdevilliers/open-telemetery-golang-bestiary/apps

go 1.14

require (
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/j2gg0s/otsql v0.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.8.0
	github.com/rs/zerolog v1.20.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.15.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.15.0
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.14.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	google.golang.org/grpc v1.34.0
)
