package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/convention"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/api/contracts"
	"github.com/nbrglm/nexeres/internal/interfaces"
	"github.com/nbrglm/nexeres/internal/mfa"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"github.com/nbrglm/nexeres/internal/models"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/password"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/utils"
	"go.opentelemetry.io/otel/metric"
)

type SignupHandler struct {
	Counter metric.Int64Counter
}

func NewSignupHandler(meter metric.Meter) (interfaces.Handler, error) {
	c, err := meter.Int64Counter("nexeres.auth.signup", metric.WithDescription("Total number of Signup requests"))
	if err != nil {
		return nil, err
	}

	return &SignupHandler{
		Counter: c,
	}, nil
}

func (h *SignupHandler) Register(r chi.Router) {
	r.With(middlewares.DecodeAndValidate[contracts.AuthSignupRequest]()).Post("/signup", h.Handle)
}

func (h *SignupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Create a new span for the request
	ctx, log, span := obs.WithContext(r.Context(), "SignupHandler.Handle")
	defer span.End()

	// Increment the request counter
	h.Counter.Add(ctx, 1, convention.OptAttrSetTotal)

	body := middlewares.GetDecodedBody[contracts.AuthSignupRequest](ctx, w)
	if body == nil {
		// Error response already sent by the middleware
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to begin transaction"), span, log, h.Counter, "SignupHandler.Handle.BeginTx")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	domain, err := utils.GetDomainFromEmail(body.Email)
	if err != nil {
		resp.ProcessError(w, ctx, contracts.BadRequest("Invalid email address"), span, log, h.Counter, "SignupHandler.Handle.GetDomainFromEmail")
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

	if config.C.Multitenancy.Enabled {
		if body.InviteToken == nil || strings.TrimSpace(*body.InviteToken) == "" {
			// Check if the domain matches any verified, auto-join enabled domains for any organization
			orgDomainVerified := true
			orgAutoJoin := true
			organization, err := q.GetOrgByDomain(ctx, db.GetOrgByDomainParams{
				Domain:   domain,
				Verified: &orgDomainVerified,
				AutoJoin: &orgAutoJoin,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				resp.ProcessError(w, ctx, contracts.Unauthorized("Email is not associated to any Organizations! Please contact your administrator.", "No organization found for domain that has auto-join enabled! If you have not verified the domain yet, please do so."), span, log, h.Counter, "SignupHandler.Handle")
				return
			}

			if err != nil {
				resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to get organization for domain!"), span, log, h.Counter, "SignupHandler.Handle")
				return
			}
			// we do not change role here, as we are not using an invite token,
			// so the user will be a member by default.
			org = &organization.Org
			roleId = organization.AutoJoinRoleID
		} else {
			invitation, err := q.GetInvitation(ctx, db.GetInvitationParams{
				TokenHash: body.InviteToken,
				Statuses:  []db.InvitationStatus{db.InvitationStatusPending},
			})
			if errors.Is(err, pgx.ErrNoRows) {
				resp.ProcessError(w, ctx, contracts.Unauthorized("Invalid invite token! Please check your token and try again.", "No invitation found for the provided token!"), span, log, h.Counter, "SignupHandler.Handle")
				return
			}
			if err != nil {
				resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to get invitation by token!"), span, log, h.Counter, "SignupHandler.Handle")
				return
			}
			if invitation.Email != body.Email {
				resp.ProcessError(w, ctx, contracts.Unauthorized("Invalid invite token! Please check your token and try again.", "The invite token does not match the provided email!"), span, log, h.Counter, "SignupHandler.Handle")
				return
			}

			// The invite is valid, let's get the organization and role from the invitation
			organization, err := q.GetOrg(ctx, db.GetOrgParams{
				ID: &invitation.OrgID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				resp.ProcessError(w, ctx, contracts.Unauthorized("Invalid invite token! Please check your token and try again.", "No organization found for the provided invitation!"), span, log, h.Counter, "SignupHandler.Handle")
				return
			}
			if err != nil {
				resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to get organization by ID for given invite token!"), span, log, h.Counter, "SignupHandler.Handle")
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
			resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to get default organization!"), span, log, h.Counter, "SignupHandler.Handle")
			return
		}
		org = &organization
	}

	userId, err := uuid.NewV7()
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to generate user ID!"), span, log, h.Counter, "SignupHandler.Handle")
		return
	}

	passwordHash, err := password.HashPassword(body.Password)
	if err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to hash password!"), span, log, h.Counter, "SignupHandler.Handle")
		return
	}

	// Create the user
	err = q.CreateUser(ctx, db.CreateUserParams{
		ID:           userId,
		Email:        body.Email,
		Name:         body.Name,
		PasswordHash: &passwordHash,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			resp.ProcessError(w, ctx, contracts.BadRequest("Email is already registered! Please login or use a different email.", "User with this email already exists!"), span, log, h.Counter, "SignupHandler.Handle")
			return
		}
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to create user!"), span, log, h.Counter, "SignupHandler.Handle")
		return
	}

	// Create the user organization membership
	if err := q.AddUserToOrg(ctx, db.AddUserToOrgParams{
		UserID:     userId,
		OrgID:      org.ID,
		RoleID:     roleId,
		IsOrgAdmin: isOrgAdmin,
	}); err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to link user to organization!"), span, log, h.Counter, "SignupHandler.Handle")
		return
	}

	var backupCodes models.BackupCodes = make(models.BackupCodes, 0)

	// MFA Check
	if org.Settings.MFA.Required {
		// Generate 8 backup codes
		backupCodes, err = mfa.GenerateBackupCodes(8)
		if err != nil {
			resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to generate MFA backup codes!"), span, log, h.Counter, "SignupHandler.Handle")
			return
		}

		// Enable MFA for the user
		if err := q.UpdateMFA(ctx, db.UpdateMFAParams{
			MfaEnabled:  true,
			BackupCodes: backupCodes,
			ID:          &userId,
		}); err != nil {
			resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to enable MFA for user!"), span, log, h.Counter, "SignupHandler.Handle")
			return
		}

		mfaFactorId, err := uuid.NewV7()
		if err != nil {
			resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to generate MFA factor ID!"), span, log, h.Counter, "SignupHandler.Handle")
			return
		}

		// Add the email factor
		if err := q.AddMFAFactor(ctx, db.AddMFAFactorParams{
			ID:         mfaFactorId,
			UserID:     userId,
			FactorType: db.MfaTypeEmail,
			Name:       "Registered Email",
			Secret:     body.Email,
			// No verified specified as we need to verify the email factor
			// when the user completes their email verification.
		}); err != nil {
			resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to add email factor for user!"), span, log, h.Counter, "SignupHandler.Handle")
			return
		}
	}

	msg := "Signup successful!"
	if org.Settings.MFA.Required {
		msg += " Multi-factor authentication (MFA) is required for your account and has been enabled. Please verify your email before logging in — this will add your email as a backup factor. Also, save your backup codes securely in case you lose access to your email."
	}

	if err := tx.Commit(ctx); err != nil {
		resp.ProcessError(w, ctx, contracts.InternalServerError(err, "Failed to commit transaction"), span, log, h.Counter, "SignupHandler.Handle.CommitTx")
		return
	}

	h.Counter.Add(ctx, 1, convention.OptAttrSetSuccess)
	resp.WriteJSON(w, http.StatusOK, contracts.AuthSignupResponse{
		BaseResponse: contracts.BaseResponse{
			Success: true,
			Message: msg,
		},
		BackupCodes: &backupCodes,
		UserID:      userId.String(),
	})
}
