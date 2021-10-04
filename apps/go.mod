module github.com/mdevilliers/open-telemetery-golang-bestiary/apps

go 1.14

require (
	cloud.google.com/go v0.78.0 // indirect
	github.com/XSAM/otelsql v0.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.3
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.25.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.23.0
	go.opentelemetry.io/contrib/instrumentation/host v0.22.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC3
	go.opentelemetry.io/otel/exporters/prometheus v0.22.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/sdk v1.0.1
	go.opentelemetry.io/otel/sdk/export/metric v0.24.0
	go.opentelemetry.io/otel/sdk/metric v0.24.0
	go.opentelemetry.io/otel/trace v1.0.1
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93 // indirect
	google.golang.org/genproto v0.0.0-20210303154014-9728d6b83eeb // indirect
	google.golang.org/grpc v1.40.0
)
