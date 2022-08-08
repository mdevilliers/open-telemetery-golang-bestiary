module github.com/mdevilliers/open-telemetery-golang-bestiary/apps

go 1.14

require (
	cloud.google.com/go v0.78.0 // indirect
	github.com/XSAM/otelsql v0.15.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.6
	github.com/prometheus/client_golang v1.13.0
	github.com/rs/zerolog v1.27.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.33.0
	go.opentelemetry.io/contrib/instrumentation/host v0.33.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.33.0
	go.opentelemetry.io/otel v1.9.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.9.0
	go.opentelemetry.io/otel/exporters/prometheus v0.31.0
	go.opentelemetry.io/otel/metric v0.31.0
	go.opentelemetry.io/otel/sdk v1.9.0
	go.opentelemetry.io/otel/sdk/metric v0.31.0
	go.opentelemetry.io/otel/trace v1.9.0
	golang.org/x/tools v0.1.7 // indirect
	google.golang.org/grpc v1.48.0
)
