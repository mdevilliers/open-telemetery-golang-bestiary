module github.com/mdevilliers/open-telemetery-golang-bestiary/apps

go 1.14

require (
	github.com/DataDog/sketches-go v0.0.1 // indirect
	github.com/XSAM/otelsql v0.2.1
	github.com/golang/protobuf v1.4.3
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/j2gg0s/otsql v0.5.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.8.0
	github.com/rs/zerolog v1.20.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.19.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.19.0
	go.opentelemetry.io/otel v0.19.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.19.0
	go.opentelemetry.io/otel/exporters/otlp v0.19.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.19.0
	go.opentelemetry.io/otel/sdk v0.19.0
	go.opentelemetry.io/otel/trace v0.19.0
	google.golang.org/grpc v1.36.0
)
