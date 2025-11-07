package handlers

import (
	"errors"
	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type RefreshTokenHandler struct {
	RefreshTokenCounter *prometheus.CounterVec
}

func NewRefreshTokenHandler() *RefreshTokenHandler {
	return &RefreshTokenHandler{
		RefreshTokenCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "refresh_token",
				Help:      "Total number of RefreshToken requests",
			}, []string{"status"},
		),
	}
}

func (h *RefreshTokenHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.RefreshTokenCounter)
	router.POST("/api/auth/refresh", h.Handle)
}

type RefreshTokenRequest struct {
	UserIP    string `json:"userIp" binding:"required,ip"`
	UserAgent string `json:"userAgent"`
}

type RefreshTokenResponse struct {
	resp.BaseResponse
	Tokens *tokens.Tokens `json:"tokens"`
}

// @Summary RefreshToken Endpoint
// @Description Handles RefreshToken requests
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Request body"
// @Param X-NEXERES-Refresh-Token header string true "Refresh token"
// @Success 200 {object} RefreshTokenResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *RefreshTokenHandler) Handle(c *gin.Context) {
	h.RefreshTokenCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "RefreshTokenHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	ipAddress := netip.MustParseAddr(req.UserIP)
	userAgent := req.UserAgent
	// we don't check if refreshToken is empty because the RequireAuth middleware ensures it's present
	refreshToken := c.GetString(middlewares.CtxRefreshToken)

	_, refreshTokenHash := tokens.HashTokens(&tokens.Tokens{
		RefreshToken: refreshToken,
	})

	session, err := q.GetSession(ctx, db.GetSessionParams{
		RefreshTokenHash: &refreshTokenHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.ProcessError(c, reserr.Unauthorized("Invalid refresh token! Please login again.", "No session found for refresh token"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
			return
		}

		utils.ProcessError(c, reserr.InternalServerError(err, "Unable to retrieve session"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}

	// We DO NOT CHECK if the token has been revoked here as:
	// The revocation is done by deleting the session from the database.
	// So if the session exists, the token is valid and not revoked.
	// Thus, if we are here, it means the session exists and the token is valid.
	// Using the method `tokens.HasTokenBeenRevoked`, which checks if no session exists for the given session ID,
	// would be redundant.

	newTokenInfo, err := q.GetInfoForSessionRefresh(ctx, db.GetInfoForSessionRefreshParams{
		UserID: session.UserID,
		OrgID:  session.OrgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug("No user or organization found for session", zap.String("sessionID", session.ID.String()))
			utils.ProcessError(c, reserr.Unauthorized("User or organization not found! Please login again.", "No user or organization found for session"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
			return
		}

		utils.ProcessError(c, reserr.InternalServerError(err, "Unable to retrieve user or organization info"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}

	newTokenPair, err := tokens.RefreshSessionTokens(session, tokens.NexeresClaims{
		OrgSlug:      newTokenInfo.OrgSlug,
		OrgName:      newTokenInfo.OrgName,
		OrgAvatarURL: newTokenInfo.OrgAvatarUrl,
		OrgId:        session.OrgID,

		Email:         newTokenInfo.UserEmail,
		EmailVerified: newTokenInfo.UserEmailVerified,
		MFAEnabled:    newTokenInfo.UserMfaEnabled,
		OrgAdmin:      newTokenInfo.UserIsOrgAdmin,
		UserName:      newTokenInfo.UserName,
		UserAvatarURL: newTokenInfo.UserAvatarUrl,
		UserOrgRole:   newTokenInfo.UserOrgRole,
		UserOrgRoleId: newTokenInfo.UserOrgRoleID,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Unable to generate new tokens"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}

	newSessionTokenHash, newRefreshTokenHash := tokens.HashTokens(newTokenPair)
	err = q.RefreshSession(ctx, db.RefreshSessionParams{
		ID:               &session.ID,
		RefreshTokenHash: newRefreshTokenHash,
		SessionTokenHash: newSessionTokenHash,
		ExpiresAt:        newTokenPair.RefreshTokenExpiry,
		IpAddress:        &ipAddress,
		UserAgent:        &userAgent,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Unable to refresh session"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.RefreshTokenCounter, "RefreshTokenHandler.Handle")
		return
	}

	h.RefreshTokenCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &RefreshTokenResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation RefreshToken completed successfully.",
		},
		Tokens: newTokenPair,
	})
}
