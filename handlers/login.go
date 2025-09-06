package handlers

import (
	"errors"
	"net/http"
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/models"
	"github.com/nbrglm/nexeres/internal/password"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/opts"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type LoginHandler struct {
	LoginCounter *prometheus.CounterVec
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{
		LoginCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "user_login_requests",
				Help:      "Total number of user login requests",
			},
			[]string{"status"},
		),
	}
}

func (h *LoginHandler) Register(engine *gin.Engine) {
	metrics.Collectors = append(metrics.Collectors, h.LoginCounter)
	engine.POST("/api/auth/login", h.HandleLogin)
}

type UserLoginData struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`

	// Optional field to store in the flow data which can be fetched by the client after login
	// This can be used to redirect the user to a specific page after login
	// or to maintain the state of the application.
	// It is recommended to validate this field on the client side to prevent open redirect vulnerabilities.
	FlowReturnTo *string `json:"flowReturnTo,omitempty"`
}

type UserLoginResult struct {
	Message                  string         `json:"message"`
	Tokens                   *tokens.Tokens `json:"tokens,omitempty"`
	RequireEmailVerification bool           `json:"requireEmailVerification"`
	FlowID                   *string        `json:"flowId,omitempty"`
}

// HandleLogin godoc
// @Summary User Login
// @Description Handles user login requests.
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body UserLoginData true "User Login Data"
// @Success 200 {object} UserLoginResult "User Login Result"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid Credentials or User does not belong to any organization"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /api/auth/login [post]
func (h *LoginHandler) HandleLogin(c *gin.Context) {
	h.LoginCounter.WithLabelValues("received").Inc()
	if config.Multitenancy {
		h.handleMultitenantLogin(c)
	} else {
		h.handleSingleTenantLogin(c)
	}
	// no return required, the functions above handle the response
}

func (h *LoginHandler) handleSingleTenantLogin(c *gin.Context) {
	// For single-tenant login, we can directly validate the user's credentials
	ctx, log, span := internal.WithContext(c.Request.Context(), "login")
	defer span.End() // Ensure the span is ended to avoid memory leaks

	var loginData UserLoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		utils.ProcessError(c, models.NewErrorResponse("Invalid input data", "Bad Request", http.StatusBadRequest, nil), span, log, h.LoginCounter, "login")
		return
	}

	tx, err := store.PgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to begin transaction!", http.StatusInternalServerError, nil), span, log, h.LoginCounter, "login")
		return
	}
	defer tx.Rollback(ctx)

	q := store.Querier.WithTx(tx)

	log.Debug("Retrieving user information")
	user, err := q.GetLoginInfoForUser(ctx, loginData.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Debug("User not found", zap.String("email", loginData.Email))
		utils.ProcessError(c, models.NewErrorResponse("Invalid email or password! Please try again.", "User not found!", http.StatusUnauthorized, nil), span, log, h.LoginCounter, "login")
		return
	}
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse("An error occurred while processing your request. Please try again later.", "Failed to retrieve user information!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}

	if !user.EmailVerified {
		log.Debug("User email not verified", zap.String("email", loginData.Email))
		c.JSON(http.StatusOK, &UserLoginResult{
			Message:                  "Please verify your email before logging in.",
			RequireEmailVerification: true,
		})
		return
	}

	log.Debug("Verifying user password")
	if !password.VerifyPasswordMatch(*user.PasswordHash, loginData.Password) {
		log.Debug("Password mismatch", zap.String("email", loginData.Email))
		utils.ProcessError(c, models.NewErrorResponse("Invalid credentials! Please try again.", "Password mismatch!", http.StatusUnauthorized, nil), span, log, h.LoginCounter, "login")
		return
	}

	avatarUrl := ""
	if user.AvatarUrl != nil {
		avatarUrl = *user.AvatarUrl
	}

	log.Debug("Generating tokens")
	tokensResult, err := tokens.GenerateTokens(user.ID, tokens.NexeresClaims{
		OrgSlug: opts.DefaultOrgSlug,
		OrgName: opts.DefaultOrgName,
		OrgId:   opts.DefaultOrgId,

		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		UserFname:     *user.FirstName,
		UserMname:     user.MiddleName,
		UserLname:     *user.LastName,
		UserAvatarURL: avatarUrl,
		UserOrgRole:   "member",
	})
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to generate token pair!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}

	newSessionTokenHash, newRefreshTokenHash := tokens.HashTokens(tokensResult)
	log.Debug("Creating session in database", zap.String("sessionId", tokensResult.SessionId.String()))

	ipAddress := netip.MustParseAddr(c.ClientIP())
	userAgent := c.Request.UserAgent()

	_, err = q.CreateSession(ctx, db.CreateSessionParams{
		ID:               tokensResult.SessionId,
		UserID:           user.ID,
		OrgID:            uuid.MustParse(opts.DefaultOrgId),
		TokenHash:        newSessionTokenHash,
		RefreshTokenHash: newRefreshTokenHash,
		MfaVerified:      false,
		MfaVerifiedAt: pgtype.Timestamptz{
			Valid: false,
		},
		IpAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: pgtype.Timestamptz{
			Time:  tokensResult.RefreshTokenExpiry,
			Valid: true,
		},
	})

	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to create session in the db!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to commit transaction!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}

	log.Debug("Login successful", zap.String("email", user.Email), zap.String("sessionId", tokensResult.SessionId.String()))
	c.JSON(http.StatusOK, &UserLoginResult{
		Message: "Login successful",
		Tokens:  tokensResult,
	})
}

func (h *LoginHandler) handleMultitenantLogin(c *gin.Context) {
	ctx, log, span := internal.WithContext(c.Request.Context(), "login")
	defer span.End() // Ensure the span is ended to avoid memory leaks

	var loginData UserLoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		utils.ProcessError(c, models.NewErrorResponse("Invalid input data", "Bad Request", http.StatusBadRequest, nil), span, log, h.LoginCounter, "login")
		return
	}

	ipAddress := netip.MustParseAddr(c.ClientIP())
	userAgent := c.Request.UserAgent()

	_, err := utils.GetDomainFromEmail(loginData.Email)
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse("Invalid request! Please input a valid email and try again.", "Invalid email domain!", http.StatusBadRequest, nil), span, log, h.LoginCounter, "login")
		return
	}

	tx, err := store.PgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to begin transaction!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}
	defer tx.Rollback(ctx)

	q := store.Querier.WithTx(tx)

	user, err := q.GetLoginInfoForUser(ctx, loginData.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		utils.ProcessError(c, models.NewErrorResponse("Invalid email or password! Please try again.", "User not found!", http.StatusUnauthorized, nil), span, log, h.LoginCounter, "login")
		return
	}
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to retrieve user information!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}

	if !user.EmailVerified {
		log.Debug("User email not verified", zap.String("email", loginData.Email))
		c.JSON(http.StatusOK, &UserLoginResult{
			Message:                  "Please verify your email before logging in.",
			RequireEmailVerification: true,
		})
		return
	}

	if !password.VerifyPasswordMatch(*user.PasswordHash, loginData.Password) {
		utils.ProcessError(c, models.NewErrorResponse("Invalid credentials! Please try again.", "Password mismatch!", http.StatusUnauthorized, nil), span, log, h.LoginCounter, "login")
		return
	}

	// Fetch the organizations the user belongs to
	orgs, err := q.GetUserOrgsByEmail(ctx, &user.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		utils.ProcessError(c, models.NewErrorResponse("You do not belong to any organization! Please contact your administrator.", "No organizations found for the user!", http.StatusUnauthorized, nil), span, log, h.LoginCounter, "login")
		return
	}
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to retrieve user organizations!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}
	if len(orgs) == 0 {
		utils.ProcessError(c, models.NewErrorResponse("You do not belong to any organization! Please contact your administrator.", "No organizations found for the user!", http.StatusUnauthorized, nil), span, log, h.LoginCounter, "login")
		return
	}
	if len(orgs) == 1 {
		avatarUrl := ""
		if user.AvatarUrl != nil {
			avatarUrl = *user.AvatarUrl
		}

		// If the user belongs to a single organization, create a session for that organization
		result, err := tokens.GenerateTokens(user.ID, tokens.NexeresClaims{
			OrgSlug: orgs[0].Org.Slug,
			OrgName: orgs[0].Org.Name,
			OrgId:   orgs[0].Org.ID.String(),

			Email:         user.Email,
			EmailVerified: user.EmailVerified,
			UserFname:     *user.FirstName,
			UserLname:     *user.LastName,
			UserAvatarURL: avatarUrl,
			UserOrgRole:   orgs[0].UserOrg.Role,
		})
		if err != nil {
			utils.ProcessError(c, models.NewErrorResponse("An error occurred while processing your request. Please try again later.", "Failed to generate token pair!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
			return
		}

		newSessionTokenHash, newRefreshTokenHash := tokens.HashTokens(result)

		_, err = q.CreateSession(ctx, db.CreateSessionParams{
			ID:               result.SessionId,
			UserID:           user.ID,
			OrgID:            orgs[0].Org.ID,
			TokenHash:        newSessionTokenHash,
			RefreshTokenHash: newRefreshTokenHash,
			MfaVerified:      false,
			MfaVerifiedAt: pgtype.Timestamptz{
				Valid: false,
			},
			IpAddress: ipAddress,
			UserAgent: userAgent,
			ExpiresAt: pgtype.Timestamptz{
				Time:  result.RefreshTokenExpiry,
				Valid: true,
			},
		})

		if err != nil {
			utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to create session in the db!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
			return
		}

		if err := tx.Commit(ctx); err != nil {
			utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to commit transaction!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
			return
		}

		c.JSON(http.StatusOK, &UserLoginResult{
			Message: "Login successful",
			Tokens:  result,
		})
		return
	}

	// If the user belongs to multiple organizations, create a flow to let the user select the organization
	organizations := make([]models.OrgCompat, len(orgs))
	for i, o := range orgs {
		organizations[i] = *models.NewOrgCompat(&o.Org)
	}

	var flow *cache.FlowData

	fId, err := uuid.NewV7()
	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to generate flow ID!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}
	flow = &cache.FlowData{
		ID:          fId.String(),
		Type:        cache.FlowTypeLogin,
		UserID:      user.ID.String(),
		Email:       user.Email,
		Orgs:        organizations,
		MFARequired: false, // MFA is not required at this stage, user has not selected an org yet
		MFAVerified: false,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // Flow expires in 10 minutes
	}

	if loginData.FlowReturnTo != nil {
		flow.ReturnTo = *loginData.FlowReturnTo
	}

	err = cache.StoreFlow(ctx, *flow)

	if err != nil {
		utils.ProcessError(c, models.NewErrorResponse(models.GenericErrorMessage, "Failed to store flow data!", http.StatusInternalServerError, err), span, log, h.LoginCounter, "login")
		return
	}

	// Return the flow ID to the client to let them select the organization
	// The client can then use this flow ID to complete the login process
	// by selecting the organization and calling the appropriate endpoint
	log.Debug("Multiple organizations found for user, returning flow ID", zap.String("flowId", flow.ID), zap.String("userEmail", user.Email))
	// Note: Do not return tokens at this stage as the user has not selected an organization
	c.JSON(http.StatusOK, &UserLoginResult{
		Message: "Multiple organizations found. Please select an organization to continue.",
		FlowID:  &flow.ID,
	})
}
