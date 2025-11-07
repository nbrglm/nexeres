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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/nbrglm/nexeres/internal/reserr"
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
func RequireAuth(mode AuthMode) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		satisfied, errStr := internalSatisfiesAuthMode(ctx, mode)
		if !satisfied {
			logging.Logger.Debug("Authentication failed in RequireAuth middleware", zap.String("mode", string(mode)), zap.String("reason", errStr))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, reserr.Unauthorized("Unauthorized access!", errStr).Filter())
			return
		}
		ctx.Next()
	}
}

// internalSatisfiesAuthMode checks if the current request context satisfies the given AuthMode.
//
// Uses recursion to evaluate combined modes.
func internalSatisfiesAuthMode(ctx *gin.Context, mode AuthMode) (bool, string) {
	_, sessionExists := ctx.Get(CtxSessionToken)
	_, refreshExists := ctx.Get(CtxRefreshToken)
	_, adminExists := ctx.Get(CtxAdminToken)

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
		if claims, exists := ctx.Get(CtxSessionTokenClaims); !exists {
			return false, "Invalid or missing session token"
		} else {
			c, ok := claims.(*tokens.NexeresClaims)
			if !ok {
				return false, "Invalid session token"
			}
			return c.OrgAdmin, "User is not an organization admin"
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

func PopulateAuthContext() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := strings.TrimSpace(ctx.GetHeader(tokens.NEXERES_API_KeyHeaderName))

		if apiKey == "" {
			logging.Logger.Warn("Missing API key in request")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, reserr.Unauthorized("Unauthorized access!", "Missing API key").Filter())
			return
		}

		// Validate the API Key
		exists := slices.ContainsFunc(config.C.Security.APIKeys, func(key config.APIKeyConfig) bool {
			if apiKey == key.Key {
				ctx.Set(CtxAPIKeyGetter, key)
				return true
			}
			return false
		})

		if !exists {
			logging.Logger.Warn("Invalid API key provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, reserr.Unauthorized("Unauthorized access!", "Invalid API key").Filter())
			return
		}

		// Session token
		sessionToken := strings.TrimSpace(ctx.GetHeader(tokens.SessionTokenHeaderName))
		if sessionToken != "" {
			bytes, err := base64.RawURLEncoding.DecodeString(sessionToken)
			if err != nil {
				logging.Logger.Debug("Failed to decode session token", zap.Error(err))
			} else {
				// Validate the session token
				claims, err := ValidateSessionToken(ctx.Request.Context(), string(bytes))
				if err != nil {
					logging.Logger.Debug("Failed to validate session token", zap.Error(err))
				} else if claims != nil {
					ctx.Set(CtxSessionToken, sessionToken)
					ctx.Set(CtxSessionTokenClaims, claims)
				}
			}
		}

		// Refresh token
		refreshToken := strings.TrimSpace(ctx.GetHeader(tokens.RefreshTokenHeaderName))
		if refreshToken != "" {
			ctx.Set(CtxRefreshToken, refreshToken)
		}

		// Admin token
		adminToken := strings.TrimSpace(ctx.GetHeader(tokens.AdminTokenHeaderName))
		if adminToken != "" {
			hash := tokens.HashAdminToken(adminToken)
			sess, err := cache.GetAdminSession(ctx.Request.Context(), hash) // Just to check if it exists
			if err != nil || sess == nil {
				if errors.Is(err, cache.ErrKeyNotFound) {
					logging.Logger.Debug("Admin token expired or not found in cache")
				} else if err != nil {
					logging.Logger.Debug("Failed to validate admin token", zap.Error(err))
				}
			} else {
				ctx.Set(CtxAdminToken, hash)
				ctx.Set(CtxAdminEmail, sess.Email)
			}
		}

		ctx.Next()
	}
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
