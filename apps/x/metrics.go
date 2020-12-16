package x

import (
	"net/http"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
)

func IntialiseMetrics() error {

	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		return err
	}
	http.HandleFunc("/metrics", exporter.ServeHTTP)
	return nil
}
