package resp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/nbrglm/nexeres/convention"
	"github.com/nbrglm/nexeres/internal/api/contracts"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// WriteJSON writes the given data as a JSON response with the specified HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ProcessError is a utility function to handle errors in a consistent way across handlers.
// It logs the error, increments the appropriate Prometheus counter, and sends a JSON response to the client.
// Returns true if an error occurred and was handled, false otherwise.
func ProcessError(w http.ResponseWriter, ctx context.Context, err *contracts.ErrorResponse, span trace.Span, log *zap.Logger, counter metric.Int64Counter, opName string) bool {
	if err == nil {
		return false
	}

	counter.Add(ctx, 1, convention.OptAttrSetError)
	log.Debug("Error occurred during operation!", zap.String("operation", opName), zap.Error(err))

	if err.UnderlyingError != nil {
		// Log and Record the underlying error if it exists
		log.Error("Failed to handle operation", zap.String("operation", opName), zap.Error(err.UnderlyingError))
		span.RecordError(err.UnderlyingError)
	}
	span.SetStatus(codes.Error, err.Debug)
	err.Write(w)
	return true
}
