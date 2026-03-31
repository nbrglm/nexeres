package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/nbrglm/nexeres/internal/interfaces"
	"github.com/nbrglm/nexeres/internal/metrics"
)

func RegisterAPIRoutes(r chi.Router) error {
	constructors := []interfaces.HandlerConstructor{
		NewSignupHandler,
		NewSendVerificationEmailHandler,
		NewVerifyEmailTokenHandler,
	}
	handlers := make([]interfaces.Handler, 0, len(constructors))
	for _, constructor := range constructors {
		handler, err := constructor(metrics.Meter)
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
	}

	r.Route("/auth", func(r chi.Router) {
		for _, handler := range handlers {
			handler.Register(r)
		}
	})
	return nil
}
