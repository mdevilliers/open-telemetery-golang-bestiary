module github.com/mdevilliers/open-telemetery-golang-bestiary/apps

go 1.14

require (
	github.com/golang/protobuf v1.4.3
	github.com/j2gg0s/otsql v0.3.0
	github.com/lib/pq v1.8.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.14.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.14.0
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
	google.golang.org/grpc v1.33.2
)
