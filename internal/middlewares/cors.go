package middlewares

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/opts"
)

func InitCORS(engine *gin.Engine) {
	engine.Use(cors.New(
		cors.Config{
			AllowCredentials: true, // We do not let the user set this, as we need credentials for the UI.
			AllowMethods:     config.C.Security.CORS.AllowedMethods,
			AllowOriginFunc:  isOriginAllowed,
			AllowHeaders:     config.C.Security.CORS.AllowedHeaders,
		},
	))
}

func isOriginAllowed(origin string) bool {
	// Since GetBaseURL() returns the debugBaseURL if opts.Debug is true,
	// thus we check that first.
	if origin == config.C.Public.GetBaseURL() {
		return true
	}

	// If the origin is not equal to the base URL, we check for debug mode.
	// If opts.Debug is true, we allow the origin to be empty.
	if opts.Debug {
		return origin == ""
	}

	if origin == "https://"+config.C.Public.Domain {
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
