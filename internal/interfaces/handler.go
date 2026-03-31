package interfaces

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/metric"
)

type Handler interface {
	Register(r chi.Router)
	Handle(w http.ResponseWriter, r *http.Request)
}

type HandlerConstructor func(metric.Meter) (Handler, error)
