package handlers

import (
	"errors"
	"maps"
	"net/http"
	"net/netip"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/password"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
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
				Name:      "login",
				Help:      "Total number of Login requests",
			}, []string{"status"},
		),
	}
}

func (h *LoginHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.LoginCounter)
	router.POST("/api/auth/login", h.Handle)
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`

	// Optional field to store in the flow data which can be fetched by the client after login
	// This can be used to redirect the user to a specific page after login
	// or to maintain the state of the application.
	// It is recommended to validate this field on the client side to prevent open redirect vulnerabilities.
	FlowReturnTo *string `json:"flowReturnTo,omitempty"`

	// Metadata, required to issue tokens
	UserIP    string `json:"userIp" binding:"required,ip"`
	UserAgent string `json:"userAgent"`
}

type LoginResponse struct {
	resp.BaseResponse
	Tokens                   *tokens.Tokens `json:"tokens,omitempty"`
	RequireEmailVerification bool           `json:"requireEmailVerification"`
	FlowID                   *string        `json:"flowId,omitempty"`
}

// @Summary Login Endpoint
// @Description Handles Login requests
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Request body"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/login [post]
func (h *LoginHandler) Handle(c *gin.Context) {
	h.LoginCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "LoginHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	ipAddress := netip.MustParseAddr(req.UserIP)
	userAgent := req.UserAgent

	user, err := q.GetLoginInfoForUser(ctx, db.GetLoginInfoForUserParams{
		Email: &req.Email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		utils.ProcessError(c, reserr.Unauthorized("Invalid email or password! Please try again.", "User not found!"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to retrieve user information!"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}

	if !password.VerifyPasswordMatch(*user.PasswordHash, req.Password) {
		utils.ProcessError(c, reserr.Unauthorized("Invalid credentials! Please try again.", "Password mismatch!"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusOK, &LoginResponse{
			BaseResponse: resp.BaseResponse{
				Success: false,
				Message: "Please verify your email before logging in.",
			},
			RequireEmailVerification: true,
		})
		return
	}

	// Fetch the organizations the user belongs to
	orgs, err := q.GetUserOrgs(ctx, db.GetUserOrgsParams{
		UserEmail: &user.Email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		utils.ProcessError(c, reserr.Unauthorized("You do not belong to any organization! Please contact your administrator.", "No organizations found for the user!"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to retrieve user organizations!"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}
	if len(orgs) == 0 {
		utils.ProcessError(c, reserr.Unauthorized("You do not belong to any organization! Please contact your administrator.", "No organizations found for the user!"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}

	// Check mfa requirement
	// We assume MFA factors exist and are verified; detailed validation happens in MFA endpoint.
	// we do not handle the condition that user's MFA is disabled
	// since we assume in the system that if a User is joining an org
	// that requires MFA, we auto-enable MFA and if the user's email is verified (which it will be since they cannot login without it),
	// their email will already be a factor, and that they have their backup codes ready.
	// We assume MFA factors exist and are verified; detailed validation happens in MFA endpoint.
	mfaRequired := user.MfaEnabled
	for _, o := range orgs {
		roles := slices.Collect(maps.Values(o.Org.Settings.MFA.Roles))
		if o.Org.Settings.MFA.Required && (len(roles) == 0 || slices.Contains(roles, o.Role.ID)) {
			mfaRequired = true
			break
		}
	}

	organizations := make([]db.Org, len(orgs))
	for i := range orgs {
		organizations[i] = orgs[i].Org
	}

	if mfaRequired || len(orgs) > 1 {
		message := "Multiple organizations found. Please select an organization to continue."

		if mfaRequired {
			message = "Multi-factor authentication (MFA) is required for your account and has been enabled. Complete the MFA verification to continue."
		}

		flowId, err := CreateLoginFlow(
			CreateLoginFlowParams{
				UserId:       user.ID.String(),
				Email:        user.Email,
				Orgs:         organizations,
				FlowReturnTo: req.FlowReturnTo,
				MFARequired:  mfaRequired,
			},
			ctx, c, span, log, h.LoginCounter, "LoginHandler.Handle",
		)
		if err != nil {
			// CreateLoginFlow already returns error responses
			// we need not return them again
			return
		}

		h.LoginCounter.WithLabelValues("success_mfa_or_orgselect").Inc()
		// Return the flow ID to the client to continue the login process
		c.JSON(http.StatusOK, &LoginResponse{
			BaseResponse: resp.BaseResponse{
				Success: true,
				Message: message,
			},
			FlowID: &flowId,
		})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.LoginCounter, "LoginHandler.Handle")
		return
	}

	// user belongs to only 1 org AND mfa is not required
	// proceed to issue tokens
	org := orgs[0]
	result, err := issueTokens(IssueTokenParams{
		Q:       q,
		User:    NewIssueTokenUserForLoginInfo(user),
		Org:     NewIssueTokenOrg(org),
		UserOrg: NewIssueTokenUserOrg(org),
		Role:    NewIssueTokenRole(org),
	}, ipAddress, userAgent, ctx, c, span, log, h.LoginCounter, "LoginHandler.Handle")
	if err != nil {
		// Issue token already returns error responses
		// we need not return them again
		return
	}

	h.LoginCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &LoginResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Login completed successfully.",
		},
		Tokens: result,
	})
}
