package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/nbrglm/nexeres/internal/api/handlers/auth"
)

func RegisterAPIRoutes(router chi.Router) {
	// Register your API routes here
	// Example:
	// router.Get("/api/example", exampleHandler)
	router.Route("/api/v1/", func(r chi.Router) {
		auth.RegisterAPIRoutes(r)
	})
}
