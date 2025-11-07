// Package metrics provides a simple interface for collecting and reporting metrics.
// It is designed to be used with the Grafana stack, using the Prometheus go client library.
//
// When a new metric is created, it is registered with the list of collectors here.
// This allows the metrics to be collected and reported easily.
package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

// The list of metrics Collectors.
//
// Any new metrics should be added to this list for them to be collected and reported.
var Collectors []prometheus.Collector

// Initialize the metrics collection system.
//
// This function should be called once at the start of the application.
func InitMetrics() {
	prometheus.MustRegister(Collectors...)
}

func RegisterCollector(collector prometheus.Collector) {
	Collectors = append(Collectors, collector)
}

// Add the metrics export route to the gin engine.
//
// The route will be available at /metrics.
// The metrics will be served in the Prometheus format.
func AddMetricsRoute(engine *gin.Engine) {
	engine.Any("/metrics", func(ctx *gin.Context) {
		handler := promhttp.Handler()
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	})
}
