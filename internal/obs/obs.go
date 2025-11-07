package obs

import (
	"context"

	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/nbrglm/nexeres/internal/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// WithContext creates a new context with a child span for tracing and returns the context, a logger with the trace ID, and the span.
// This function is useful for adding tracing information to logs and spans in a request context.
//
// NOTE: You need to call `span.End()` when you are done with the span to ensure proper cleanup and reporting of the span.
func WithContext(ctx context.Context, spanName string) (context.Context, *zap.Logger, trace.Span) {
	c, span := tracing.Tracer.Start(ctx, spanName)
	return c, logging.Logger.With(zap.Any("context", c)), span
}
