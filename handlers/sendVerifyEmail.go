package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/notifications"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type SendVerifyEmailHandler struct {
	SendVerifyEmailCounter *prometheus.CounterVec
}

func NewSendVerifyEmailHandler() *SendVerifyEmailHandler {
	return &SendVerifyEmailHandler{
		SendVerifyEmailCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "send_verify_email",
				Help:      "Total number of SendVerifyEmail requests",
			}, []string{"status"},
		),
	}
}

func (h *SendVerifyEmailHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SendVerifyEmailCounter)
	router.POST("/api/auth/verify-email/send", h.Handle)
}

type SendVerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SendVerifyEmailResponse struct {
	resp.BaseResponse
}

// @Summary SendVerifyEmail Endpoint
// @Description Handles SendVerifyEmail requests
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SendVerifyEmailRequest true "Request body"
// @Success 200 {object} SendVerifyEmailResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/verify-email/send [post]
func (h *SendVerifyEmailHandler) Handle(c *gin.Context) {
	h.SendVerifyEmailCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "SendVerifyEmailHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	// Check if the user exists
	user, err := q.GetUser(ctx, db.GetUserParams{
		Email: &req.Email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		utils.ProcessError(c, reserr.BadRequest("It seems the user with the provided email does not exist! Please check the email and try again.", "No user exists with the provided email address."), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to retrieve user information!"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	if user.EmailVerified {
		utils.ProcessError(c, reserr.BadRequest("The email is already verified!", "Email already verified!"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	// Generate a verification token
	token, hash, err := tokens.GenerateEmailVerificationToken()
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate verification token!"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	// Generate the token ID
	tokenId, err := uuid.NewV7()
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate token ID!"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	// Insert the verification token into the database
	err = q.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		ID:        tokenId,
		UserID:    user.ID,
		TokenType: db.VerificationTokenTypeEmailVerification,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to insert verification token into the database!"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	// Send the verification email
	if err := notifications.SendWelcomeEmail(ctx, notifications.SendWelcomeEmailParams{
		User: struct {
			Email string
			Name  string
		}{
			Email: user.Email,
			Name:  user.Name,
		},
		VerificationToken: token,
		ExpiresAt:         expiresAt,
	}); err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to send verification email!"), span, log, h.SendVerifyEmailCounter, "SendVerifyEmailHandler.Handle")
		return
	}

	h.SendVerifyEmailCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SendVerifyEmailResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Send Verification Email completed successfully.",
		},
	})
}
