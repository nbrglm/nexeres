package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type VerifyEmailTokenHandler struct {
	VerifyEmailTokenCounter *prometheus.CounterVec
}

func NewVerifyEmailTokenHandler() *VerifyEmailTokenHandler {
	return &VerifyEmailTokenHandler{
		VerifyEmailTokenCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "verify_email_token",
				Help:      "Total number of VerifyEmailToken requests",
			}, []string{"status"},
		),
	}
}

func (h *VerifyEmailTokenHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.VerifyEmailTokenCounter)
	router.POST("/api/auth/verify-email/verify", h.Handle)
}

type VerifyEmailTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type VerifyEmailTokenResponse struct {
	resp.BaseResponse
}

// @Summary VerifyEmailToken Endpoint
// @Description Handles VerifyEmailToken requests
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailTokenRequest true "Request body"
// @Success 200 {object} VerifyEmailTokenResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/verify-email/verify [post]
func (h *VerifyEmailTokenHandler) Handle(c *gin.Context) {
	h.VerifyEmailTokenCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "VerifyEmailTokenHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req VerifyEmailTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	// Hash the token first
	hash := tokens.HashEmailVerificationToken(req.Token)
	// Fetch the verification token from the database
	token, err := q.GetVerificationToken(ctx, db.GetVerificationTokenParams{
		TokenHash: hash,
		TokenType: db.VerificationTokenTypeEmailVerification,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		utils.ProcessError(c, reserr.BadRequest("Invalid token! Please check the token and try again.", "No verification token found with the provided token hash."), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to retrieve verification token!"), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}
	if token.TokenType != db.VerificationTokenTypeEmailVerification {
		utils.ProcessError(c, reserr.BadRequest("Invalid token! Please check the token and try again.", "The provided token is not a valid email verification token."), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}
	if token.ExpiresAt.Before(time.Now()) {
		utils.ProcessError(c, reserr.BadRequest("The token has expired! Please request a new verification email.", "The provided token has expired."), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}

	// Mark the user's email as verified, since a valid token was found
	err = q.MarkUserEmailVerified(ctx, db.MarkUserEmailVerifiedParams{
		ID: &token.UserID,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to mark user email as verified!"), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.VerifyEmailTokenCounter, "VerifyEmailTokenHandler.Handle")
		return
	}

	h.VerifyEmailTokenCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &VerifyEmailTokenResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Verify Email completed successfully.",
		},
	})
}
