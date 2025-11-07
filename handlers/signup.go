package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/mfa"
	"github.com/nbrglm/nexeres/internal/models"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/password"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type SignupHandler struct {
	SignupCounter *prometheus.CounterVec
}

func NewSignupHandler() *SignupHandler {
	return &SignupHandler{
		SignupCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "signup",
				Help:      "Total number of Signup requests",
			}, []string{"status"},
		),
	}
}

func (h *SignupHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SignupCounter)
	router.POST("/api/auth/signup", h.Handle)
}

type SignupRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	Name            string `json:"name" binding:"required,min=2"`
	InviteToken     string `json:"inviteToken,omitempty"` // Optional invite token for multitenant signup
}

type SignupResponse struct {
	resp.BaseResponse
	UserID      string             `json:"userId"`
	BackupCodes models.BackupCodes `json:"backupCodes,omitempty"`
}

// @Summary Signup Endpoint
// @Description Handles Signup requests
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignupRequest true "Request body"
// @Success 200 {object} SignupResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/signup [post]
func (h *SignupHandler) Handle(c *gin.Context) {
	h.SignupCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "SignupHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	domain, err := utils.GetDomainFromEmail(req.Email)
	if err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	// If multitenancy is enabled, we will use the invite token to determine the organization,
	// if not present, then we will check the domain against the existing organizations,
	// if the domain matches a valid, auto-join enabled, verified existing domain, we will use that organization.
	// Otherwise if multitenancy is not enabled, we will fetch the default organization.

	var org *db.Org
	var roleId *uuid.UUID
	var isOrgAdmin *bool = new(bool)
	*isOrgAdmin = false

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	if config.C.Multitenancy {
		if strings.TrimSpace(req.InviteToken) == "" {
			// Check if the domain matches any verified, auto-join enabled domains for any organization
			orgDomainVerified := true
			orgAutoJoin := true
			organization, err := q.GetOrgByDomain(ctx, db.GetOrgByDomainParams{
				Domain:   domain,
				Verified: &orgDomainVerified,
				AutoJoin: &orgAutoJoin,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				utils.ProcessError(c, reserr.Unauthorized("Email is not associated to any Organizations! Please contact your administrator.", "No organization found for domain that has auto-join enabled! If you have not verified the domain yet, please do so."), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}

			if err != nil {
				utils.ProcessError(c, reserr.InternalServerError(err, "Failed to get organization for domain!"), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}
			// we do not change role here, as we are not using an invite token,
			// so the user will be a member by default.
			org = &organization.Org
			roleId = organization.AutoJoinRoleID
		} else {
			invitation, err := q.GetInvitation(ctx, db.GetInvitationParams{
				TokenHash: &req.InviteToken,
				Statuses:  []db.InvitationStatus{db.InvitationStatusPending},
			})
			if errors.Is(err, pgx.ErrNoRows) {
				utils.ProcessError(c, reserr.Unauthorized("Invalid invite token! Please check your token and try again.", "No invitation found for the provided token!"), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}
			if err != nil {
				utils.ProcessError(c, reserr.InternalServerError(err, "Failed to get invitation by token!"), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}
			if invitation.Email != req.Email {
				utils.ProcessError(c, reserr.Unauthorized("Invalid invite token! Please check your token and try again.", "The invite token does not match the provided email!"), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}

			// The invite is valid, let's get the organization and role from the invitation
			organization, err := q.GetOrg(ctx, db.GetOrgParams{
				ID: &invitation.OrgID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				utils.ProcessError(c, reserr.Unauthorized("Invalid invite token! Please check your token and try again.", "No organization found for the provided invitation!"), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}
			if err != nil {
				utils.ProcessError(c, reserr.InternalServerError(err, "Failed to get organization by ID for given invite token!"), span, log, h.SignupCounter, "SignupHandler.Handle")
				return
			}
			roleId = invitation.RoleID // Use the role from the invitation
			*isOrgAdmin = invitation.InviteAsAdmin
			org = &organization
		}
	} else {
		slug := "default"
		organization, err := q.GetOrg(ctx, db.GetOrgParams{
			Slug: &slug,
		})
		if err != nil {
			// we return the underlying error here because default org is ALWAYS supposed to be found
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to get default organization!"), span, log, h.SignupCounter, "SignupHandler.Handle")
			return
		}
		org = &organization
	}

	userId, err := uuid.NewV7()
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate user ID!"), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	passwordHash, err := password.HashPassword(req.Password)
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to hash password!"), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	// Create the user
	err = q.CreateUser(ctx, db.CreateUserParams{
		ID:           userId,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: &passwordHash,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			utils.ProcessError(c, reserr.BadRequest("Email is already registered! Please login or use a different email.", "User with this email already exists!"), span, log, h.SignupCounter, "SignupHandler.Handle")
			return
		}
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to create user!"), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	// Create the user organization membership
	if err := q.AddUserToOrg(ctx, db.AddUserToOrgParams{
		UserID:     userId,
		OrgID:      org.ID,
		RoleID:     roleId,
		IsOrgAdmin: isOrgAdmin,
	}); err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to link user to organization!"), span, log, h.SignupCounter, "SignupHandler.Handle")
		return
	}

	var backupCodes models.BackupCodes = make(models.BackupCodes, 0)

	// MFA Check
	if org.Settings.MFA.Required {
		// Generate 8 backup codes
		backupCodes, err = mfa.GenerateBackupCodes(8)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate MFA backup codes!"), span, log, h.SignupCounter, "SignupHandler.Handle")
			return
		}

		// Enable MFA for the user
		if err := q.UpdateMFA(ctx, db.UpdateMFAParams{
			MfaEnabled:  true,
			BackupCodes: backupCodes,
			ID:          &userId,
		}); err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to enable MFA for user!"), span, log, h.SignupCounter, "SignupHandler.Handle")
			return
		}

		mfaFactorId, err := uuid.NewV7()
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate MFA factor ID!"), span, log, h.SignupCounter, "SignupHandler.Handle")
			return
		}

		// Add the email factor
		if err := q.AddMFAFactor(ctx, db.AddMFAFactorParams{
			ID:         mfaFactorId,
			UserID:     userId,
			FactorType: db.MfaTypeEmail,
			Name:       "Registered Email",
			Secret:     req.Email,
			// No verified specified as we need to verify the email factor
			// when the user completes their email verification.
		}); err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to add email factor for user!"), span, log, h.SignupCounter, "SignupHandler.Handle")
			return
		}
	}

	msg := "Signup successful!"
	if org.Settings.MFA.Required {
		msg += " Multi-factor authentication (MFA) is required for your account and has been enabled. Please verify your email before logging in — this will add your email as a backup factor. Also, save your backup codes securely in case you lose access to your email."
	}

	h.SignupCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SignupResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: msg,
		},
		BackupCodes: backupCodes,
		UserID:      userId.String(),
	})
}
