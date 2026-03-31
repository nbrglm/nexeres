package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nbrglm/nexeres/internal/api/contracts"
	"github.com/nbrglm/nexeres/internal/interfaces"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"go.opentelemetry.io/otel/metric"
)

type VerifyEmailTokenHandler struct {
	Counter metric.Int64Counter
}

func NewVerifyEmailTokenHandler(meter metric.Meter) (interfaces.Handler, error) {
	c, err := meter.Int64Counter("nexeres.auth.verify_email_token", metric.WithDescription("Total number of VerifyEmailToken requests"))
	if err != nil {
		return nil, err
	}

	return &VerifyEmailTokenHandler{
		Counter: c,
	}, nil
}

func (h *VerifyEmailTokenHandler) Register(r chi.Router) {
	r.With(middlewares.DecodeAndValidate[contracts.VerifyEmailTokenRequest]()).Post("/verify-email/verify", h.Handle)
}

func (h *VerifyEmailTokenHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Implementation goes here
}
