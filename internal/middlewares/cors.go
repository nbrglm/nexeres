package middlewares

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/convention"
	"github.com/nbrglm/nexeres/opts"
)

func InitCORS(r chi.Router) {
	r.Use(cors.Handler(
		cors.Options{
			AllowCredentials: true, // We do not let the user set this, as we need credentials for the UI.
			AllowedMethods:   convention.CORSAllowedMethods,
			AllowOriginFunc:  isOriginAllowed,
			AllowedHeaders:   convention.CORSAllowedHeaders,
		},
	))
}

func isOriginAllowed(_ *http.Request, origin string) bool {
	// Since GetBaseURL() returns the debugBaseURL if opts.Debug is true,
	// thus we check that first.
	if origin == config.C.PublicEndpoint.GetBaseURL() {
		return true
	}

	// If the origin is not equal to the base URL, we check for debug mode.
	// If opts.Debug is true, we allow the origin to be empty.
	if opts.Debug {
		return origin == ""
	}

	if origin == "https://"+config.C.PublicEndpoint.Domain {
		return true
	}

	if checkDomain(origin) {
		return true
	}

	return false
}

func checkDomain(origin string) bool {
	for _, allowedOrigin := range config.C.Security.CORS.AllowedOrigins {
		if strings.HasPrefix(allowedOrigin, "https://*.") {
			// Retain the dot, hence we use 9 instead of 8
			domain := allowedOrigin[9:]

			if strings.HasSuffix(origin, domain) {
				return true
			}
		} else {

			// Exact match only
			if origin == allowedOrigin {
				return true
			}
		}
	}

	return false
}
