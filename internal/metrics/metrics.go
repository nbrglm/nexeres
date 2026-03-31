// Package metrics provides a simple interface for collecting and reporting metrics.
// It is designed to be used with the Grafana stack, using the Prometheus go client library.
//
// When a new metric is created, it is registered with the list of collectors here.
// This allows the metrics to be collected and reported easily.
package metrics

import (
	"context"
	"os"
	"time"

	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/opts"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// The list of metrics Collectors.
//
// Any new metrics should be added to this list for them to be collected and reported.
var Collectors []prometheus.Collector

// Otel Meter
var Meter metric.Meter

// Meter provider
var MeterProvider *sdkmetric.MeterProvider

// Initialize the metrics collection system.
//
// This function should be called once at the start of the application.
func InitMetrics() error {
	prometheus.MustRegister(Collectors...)

	var exporter sdkmetric.Exporter
	var err error
	switch config.C.Observability.Metrics.Protocol {
	case "grpc":
		options := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(config.C.Observability.Metrics.Endpoint),
		}
		if config.C.Observability.Metrics.WithInsecure {
			options = append(options, otlpmetricgrpc.WithInsecure())
		}
		if len(config.C.Observability.Metrics.Headers) > 0 {
			options = append(options, otlpmetricgrpc.WithHeaders(config.C.Observability.Metrics.Headers))
		}
		exporter, err = otlpmetricgrpc.New(context.Background(), options...)
	case "http/protobuf":
		options := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(config.C.Observability.Metrics.Endpoint),
		}
		if config.C.Observability.Metrics.WithInsecure {
			options = append(options, otlpmetrichttp.WithInsecure())
		}
		if len(config.C.Observability.Metrics.Headers) > 0 {
			options = append(options, otlpmetrichttp.WithHeaders(config.C.Observability.Metrics.Headers))
		}
		if config.C.Observability.Metrics.Path != nil {
			options = append(options, otlpmetrichttp.WithURLPath(*config.C.Observability.Metrics.Path))
		}
		exporter, err = otlpmetrichttp.New(context.Background(), options...)
	case "stdout":
		exporter, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	}
	if err != nil {
		return err
	}

	hostname, _ := os.Hostname()

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(opts.Name),
		semconv.ServiceVersion(opts.Version),
		semconv.ServiceInstanceID(config.C.Server.InstanceId),
		semconv.DeploymentEnvironment(config.Environment()),
		semconv.HostName(hostname),
	)

	MeterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(time.Second*time.Duration(config.C.Observability.Metrics.CollectionIntervalSeconds)))), sdkmetric.WithResource(res))
	otel.SetMeterProvider(MeterProvider)
	Meter = MeterProvider.Meter(opts.Name)
	return nil
}

func ShutdownMetrics(ctx context.Context) error {
	if MeterProvider == nil {
		return nil
	}
	return MeterProvider.Shutdown(ctx)
}

/// TODO: Add exporter instead of prometheus HTTP handler.
