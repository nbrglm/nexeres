package middlewares

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/api/contracts"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/nbrglm/nexeres/internal/tokens"
	"go.uber.org/zap"
)

const CtxAPIKeyGetter = "apiKey"
const CtxSessionToken = "sessionToken"
const CtxRefreshToken = "refreshToken"
const CtxSessionTokenClaims = "sessionTokenClaims"
const CtxSessionRefreshTokenKey = "refreshToken"
const CtxAdminToken = "adminToken"
const CtxAdminEmail = "adminEmail"

type AuthMode string

const AuthModeEitherSessionOrRefresh AuthMode = "either"
const AuthModeSession AuthMode = "session"
const AuthModeRefresh AuthMode = "refresh"
const AuthModeBothSessionAndRefresh AuthMode = "both"
const AuthModeSysAdmin AuthMode = "sysAdmin"
const AuthModeOrgAdmin AuthMode = "orgAdmin"
const AuthModeAnyAdmin AuthMode = "anyAdmin"

// RequireAuth is a middleware that checks for the presence of authentication tokens
// based on the specified AuthMode.
func RequireAuth(mode AuthMode) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			satisfied, errStr := internalSatisfiesAuthMode(r.Context(), mode)
			if !satisfied {
				logging.Logger.Debug("Authentication failed in RequireAuth middleware", zap.String("mode", string(mode)), zap.String("reason", errStr))
				contracts.Unauthorized("Unauthorized access!", errStr).Write(w)
				return
			}
			// Call the next handler if authentication is satisfied
			next.ServeHTTP(w, r)
		})
	}
}

// internalSatisfiesAuthMode checks if the current request context satisfies the given AuthMode.
//
// Uses recursion to evaluate combined modes.
func internalSatisfiesAuthMode(ctx context.Context, mode AuthMode) (bool, string) {
	_, sessionExists := ctx.Value(CtxSessionToken).(string)
	_, refreshExists := ctx.Value(CtxRefreshToken).(string)
	_, adminExists := ctx.Value(CtxAdminToken).(string)

	switch mode {
	case AuthModeSession:
		return sessionExists, "Invalid or missing session token"
	case AuthModeRefresh:
		return refreshExists, "Invalid or missing refresh token"
	case AuthModeSysAdmin:
		return adminExists, "Invalid or missing admin token"
	case AuthModeOrgAdmin:
		if !sessionExists {
			return false, "Invalid or missing session token"
		}
		if claims, exists := ctx.Value(CtxSessionTokenClaims).(*tokens.NexeresClaims); !exists {
			return false, "Invalid or missing session token"
		} else {
			return claims.OrgAdmin, "User is not an organization admin"
		}
	case AuthModeEitherSessionOrRefresh:
		if sessionExists || refreshExists {
			return true, ""
		}
		return false, "Invalid or missing session or refresh tokens"
	case AuthModeBothSessionAndRefresh:
		if !sessionExists || !refreshExists {
			return false, "Invalid or missing session and refresh tokens"
		}
		return true, ""
	case AuthModeAnyAdmin:
		sysAdminValid, _ := internalSatisfiesAuthMode(ctx, AuthModeSysAdmin)
		orgAdminValid, _ := internalSatisfiesAuthMode(ctx, AuthModeOrgAdmin)
		if sysAdminValid || orgAdminValid {
			return true, ""
		}
		return false, "User is not an admin"
	default:
		return false, "Invalid request"
	}
}

func PopulateAuthContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// prepare a context
		ctx := r.Context()

		if config.C.Debug {
			// In debug mode, allow docs endpoint without authentication
			if strings.HasPrefix(r.URL.Path, "/docs") || strings.HasPrefix(r.URL.Path, "/openapi/") {
				next.ServeHTTP(w, r)
				return
			}
		}

		if strings.TrimSpace(r.URL.Path) == "/" {
			next.ServeHTTP(w, r)
		}

		// Api Key section
		apiKey := strings.TrimSpace(r.Header.Get(tokens.NEXERES_API_KeyHeaderName))
		if apiKey == "" {
			logging.Logger.Info("Missing API key in request")
			contracts.Unauthorized("Unauthorized access!", "Missing API key").Write(w)
			return
		}

		// Validate the API Key
		exists := slices.ContainsFunc(config.C.Security.APIKeys, func(key config.APIKeyConfig) bool {
			if apiKey == key.Key {
				ctx = context.WithValue(ctx, CtxAPIKeyGetter, key)
				return true
			}
			return false
		})

		if !exists {
			logging.Logger.Info("Invalid API key provided")
			contracts.Unauthorized("Unauthorized access!", "Invalid API key").Write(w)
			return
		}

		// Session token section
		sessionToken := strings.TrimSpace(r.Header.Get(tokens.SessionTokenHeaderName))
		if sessionToken != "" {
			bytes, err := base64.RawURLEncoding.DecodeString(sessionToken)
			if err != nil {
				logging.Logger.Debug("Failed to decode session token", zap.Error(err))
			} else {
				// Validate the session token
				claims, err := ValidateSessionToken(ctx, string(bytes))
				if err != nil {
					logging.Logger.Debug("Failed to validate session token", zap.Error(err))
				} else if claims != nil {
					ctx = context.WithValue(ctx, CtxSessionToken, sessionToken)
					ctx = context.WithValue(ctx, CtxSessionTokenClaims, claims)
				}
			}
		}

		// Refresh token section
		refreshToken := strings.TrimSpace(r.Header.Get(tokens.RefreshTokenHeaderName))
		if refreshToken != "" {
			ctx = context.WithValue(ctx, CtxRefreshToken, refreshToken)
		}

		// Admin token section
		adminToken := strings.TrimSpace(r.Header.Get(tokens.AdminTokenHeaderName))
		if adminToken != "" {
			hash := tokens.HashAdminToken(adminToken)
			sess, err := cache.GetAdminSession(ctx, hash) // Just to check if it exists
			if err != nil || sess == nil {
				if errors.Is(err, cache.ErrKeyNotFound) {
					logging.Logger.Debug("Admin token expired or not found in cache")
				} else if err != nil {
					logging.Logger.Debug("Failed to validate admin token", zap.Error(err))
				}
			} else {
				ctx = context.WithValue(ctx, CtxAdminToken, hash)
				ctx = context.WithValue(ctx, CtxAdminEmail, sess.Email)
			}
		}

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateSessionToken validates the provided session token and returns the claims if valid.
// It returns an error if the token is invalid or if there is an issue during validation.
func ValidateSessionToken(ctx context.Context, token string) (claims *tokens.NexeresClaims, err error) {
	// Always pass a POINTER TO THE CLAIMS STRUCT to jwt.ParseWithClaims
	// so that it can populate the claims with the parsed token data.
	// Passing a struct value will give errors like "cannot unmarshal ... into Go value of type jwt.Claims"
	// Since jwt.Claims is an interface, we need to use a pointer to a concrete type that implements it.
	c := &tokens.NexeresClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (interface{}, error) {
		if t.Method == jwt.SigningMethodRS256 {
			return tokens.PublicKey, nil
		} else {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(time.Minute*5))
	if parsedToken != nil {
		if v, ok := parsedToken.Claims.(*tokens.NexeresClaims); ok && parsedToken.Valid {
			return v, nil
		}
	}
	return nil, fmt.Errorf("invalid session token: %w", err)
}
