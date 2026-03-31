// Package logging provides a simple logging interface for the application.
//
// It uses the zap logger from Uber to provide structured logging.
// Otel logging exporter is used to send logs to the OpenTelemetry collector.
// It supports both gRPC and HTTP protocols for the exporter.
// Otherwise, it defaults to using the stdout exporter with pretty print.
package logging

import (
	"context"
	"fmt"
	"os"

	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/opts"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

var Provider *log.LoggerProvider

// Innitialize the Zap logger with Otel logging exporter.
//
// It configures the logging options, sets the following fields:
//
// - Time format
// - Service name
// - Service version
// - Service instance ID
// - Deployment environment
//
// If ExporterProtocol is "grpc", it uses the OTLP gRPC exporter.
// If ExporterProtocol is "http/protobuf", it uses the OTLP HTTP exporter.
// If ExporterProtocol is not set, it defaults to using the stdout exporter with pretty print,
// which initializes the logger in debug mode.
//
// NOTE: This will return nil logger provider if the protocol is "stdout", as it uses the debug logger instead.
// If the protocol is not recognized, it returns an error.
func InitLogger() (err error) {
	// If stdout is the protocol or logs export is disabled, then use debug logger.
	if config.C.Observability.Logs.Protocol == "stdout" || config.C.Observability.Logs.Endpoint == "" {
		ReplaceWithDebugLogger()
		return nil
	}

	// The exporter is the component that will send log records to the configured endpoint.
	// It can be either gRPC or HTTP, depending on the configuration.
	var exporter log.Exporter
	switch config.C.Observability.Logs.Protocol {
	case "grpc":
		options := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(config.C.Observability.Logs.Endpoint),
		}
		if config.C.Observability.Logs.WithInsecure {
			options = append(options, otlploggrpc.WithInsecure()) // Use insecure connection in debug mode.
		}
		if len(config.C.Observability.Logs.Headers) > 0 {
			options = append(options, otlploggrpc.WithHeaders(config.C.Observability.Logs.Headers))
		}
		exporter, err = otlploggrpc.New(context.Background(), options...)
	case "http/protobuf":
		options := []otlploghttp.Option{
			otlploghttp.WithEndpoint(config.C.Observability.Logs.Endpoint),
		}
		if config.C.Observability.Logs.WithInsecure {
			options = append(options, otlploghttp.WithInsecure()) // Use insecure connection in debug mode.
		}
		if config.C.Observability.Logs.Path != nil {
			options = append(options, otlploghttp.WithURLPath(*config.C.Observability.Logs.Path))
		}
		if len(config.C.Observability.Logs.Headers) > 0 {
			options = append(options, otlploghttp.WithHeaders(config.C.Observability.Logs.Headers))
		}
		exporter, err = otlploghttp.New(context.Background(), options...)
	default:
		return fmt.Errorf("unknown otel log protocol: %s", config.C.Observability.Logs.Protocol)
	}

	if err != nil {
		return fmt.Errorf("failed to create otel log exporter: %w", err)
	}

	// Create a batch processor for the exporter.
	// The batch processor will handle the batching of log records before sending them to the exporter.
	var processor log.Processor
	if config.C.Observability.Logs.Batch {
		processor = log.NewBatchProcessor(exporter)
	} else {
		processor = log.NewSimpleProcessor(exporter)
	}
	hostname, _ := os.Hostname()

	// Resource is the global resource that will be used to create loggers.
	// It includes the service name, version, instance ID, and deployment environment.
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(opts.Name),
		semconv.ServiceVersion(opts.Version),
		semconv.ServiceInstanceID(config.C.Server.InstanceId),
		semconv.DeploymentEnvironment(config.Environment()),
		semconv.HostName(hostname),
	)

	// Provider is the global logger provider that will be used to create loggers.
	Provider = log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(res),
	)

	zapConfig := zap.NewProductionConfig()

	level, err := zapcore.ParseLevel(config.C.Observability.Logs.Level)
	if err != nil {
		return err
	}
	fmt.Printf("Setting log level to %v\n", level)
	zapConfig.Level.SetLevel(level)

	// DO NOT USE RFC3339 TIME FORMAT. USE THE DEFAULT UNIX SECONDS TIME FORMAT IN ZAP.
	// zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	Logger = zap.Must(zapConfig.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return otelzap.NewCore(opts.FullName, otelzap.WithLoggerProvider(Provider))
	})))
	zap.ReplaceGlobals(Logger)
	return nil
}

// ReplaceWithDebugLogger replaces the global logger (in zap and in this package) with a new debug mode Logger.
func ReplaceWithDebugLogger() {
	Logger = zap.Must(zap.NewDevelopment())
	zap.ReplaceGlobals(Logger)
}

// ShutdownLogger flushes any buffered log entries and shuts down the logger provider if it exists.
func ShutdownLogger(ctx context.Context) error {
	if Logger != nil {
		if err := Logger.Sync(); err != nil {
			return fmt.Errorf("failed to sync logger: %w", err)
		}
	}

	// If the provider is nil, we don't need to shutdown anything.
	if Provider != nil {
		return Provider.Shutdown(ctx)
	}
	return nil
}
