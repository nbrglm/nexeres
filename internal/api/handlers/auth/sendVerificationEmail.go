package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nbrglm/nexeres/convention"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/api/contracts"
	"github.com/nbrglm/nexeres/internal/interfaces"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"github.com/nbrglm/nexeres/internal/notifications"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"go.opentelemetry.io/otel/metric"
)

type SendVerificationEmailHandler struct {
	Counter metric.Int64Counter
}

func NewSendVerificationEmailHandler(meter metric.Meter) (interfaces.Handler, error) {
	c, err := meter.Int64Counter("nexeres.auth.send_verification_email", metric.WithDescription("Total number of SendVerificationEmail requests"))
	if err != nil {
		return nil, err
	}

	return &SendVerificationEmailHandler{
		Counter: c,
	}, nil
}

func (h *SendVerificationEmailHandler) Register(r chi.Router) {
	r.With(middlewares.DecodeAndValidate[contracts.SendVerificationEmailRequest]()).Post("/verify-email/send", h.Handle)
}

func (h *SendVerificationEmailHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Create a new span for the request
	ctx, log, span := obs.WithContext(r.Context(), "SendVerificationEmailHandler.Handle")
	defer span.End()

	// Increment the request counter
	h.Counter.Add(ctx, 1, convention.OptAttrSetTotal)

	body := middlewares.GetDecodedBody[contracts.SendVerificationEmailRequest](ctx, w)
	if body == nil {
		// Error response already sent by the middleware
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to begin transaction"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.BeginTx")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	user, err := q.GetUser(ctx, db.GetUserParams{
		Email: &body.Email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		resp.ProcessError(w, ctx, contracts.BadRequest("No user found with the provided email", "Invalid credentials!"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.GetUser.NoRows")
		return
	}
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to retrieve user by email"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.GetUser")
		return
	}

	if user.EmailVerified {
		resp.ProcessError(w, ctx, contracts.BadRequest("Email is already verified", "The provided email address has already been verified."), span, log, h.Counter, "SendVerificationEmailHandler.Handle.EmailAlreadyVerified")
		return
	}

	// generate a verification token

	token, hash, err := tokens.GenerateEmailVerificationToken()
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to generate email verification token"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.GenerateToken")
		return
	}

	tokenId, err := uuid.NewV7()
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to generate token ID"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.GenerateTokenID")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	err = q.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		ID:        tokenId,
		UserID:    user.ID,
		TokenType: db.VerificationTokenTypeEmailVerification,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to store verification token"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.StoreToken")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to commit transaction"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.CommitTx")
		return
	}

	// send verification email
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
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to send verification email"), span, log, h.Counter, "SendVerificationEmailHandler.Handle.SendEmail")
		return
	}

	h.Counter.Add(ctx, 1, convention.OptAttrSetSuccess)
	resp.WriteJSON(w, http.StatusOK, contracts.SendVerificationEmailResponse{
		BaseResponse: contracts.BaseResponse{
			Success: true,
			Message: "Verification email sent successfully",
		},
	})
}
