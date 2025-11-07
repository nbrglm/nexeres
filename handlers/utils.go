package handlers

import (
	"context"
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// THIS FILE IS NOT FOR A HANDLER
// It contains utility functions used by handlers
// which cannot be placed in the utils package due to
// circular dependency issues.

type IssueTokenUser struct {
	Email         string
	EmailVerified bool
	MfaEnabled    bool
	ID            uuid.UUID
	Name          string
	AvatarUrl     *string
}

func NewIssueTokenUserForLoginInfo(user db.GetLoginInfoForUserRow) IssueTokenUser {
	return IssueTokenUser{
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		MfaEnabled:    user.MfaEnabled,
		ID:            user.ID,
		Name:          user.Name,
		AvatarUrl:     user.AvatarUrl,
	}
}

type IssueTokenOrg struct {
	Slug      string
	Name      string
	AvatarUrl *string
	ID        uuid.UUID
}

func NewIssueTokenOrg(org db.GetUserOrgsRow) IssueTokenOrg {
	return IssueTokenOrg{
		Slug:      org.Org.Slug,
		Name:      org.Org.Name,
		AvatarUrl: org.Org.AvatarUrl,
		ID:        org.Org.ID,
	}
}

type IssueTokenUserOrg struct {
	IsOrgAdmin bool
}

func NewIssueTokenUserOrg(userOrg db.GetUserOrgsRow) IssueTokenUserOrg {
	return IssueTokenUserOrg{
		IsOrgAdmin: userOrg.UserOrg.IsOrgAdmin,
	}
}

type IssueTokenRole struct {
	ID       uuid.UUID
	RoleName string
}

func NewIssueTokenRole(role db.GetUserOrgsRow) IssueTokenRole {
	return IssueTokenRole{
		ID:       role.Role.ID,
		RoleName: role.Role.RoleName,
	}
}

type IssueTokenParams struct {
	Q       db.Querier
	User    IssueTokenUser
	Org     IssueTokenOrg
	UserOrg IssueTokenUserOrg
	Role    IssueTokenRole
}

// issueTokens issues new tokens, and saves them in a session in the database.
//
// NOTE: You need to close the transaction in the calling function. This function does not close any TXs.
func issueTokens(params IssueTokenParams, ipAddress netip.Addr, userAgent string, ctx context.Context, c *gin.Context, span trace.Span, log *zap.Logger, counter *prometheus.CounterVec, opName string) (*tokens.Tokens, error) {
	result, err := tokens.GenerateTokens(params.User.ID, tokens.NexeresClaims{
		OrgSlug:      params.Org.Slug,
		OrgName:      params.Org.Name,
		OrgAvatarURL: params.Org.AvatarUrl,
		OrgId:        params.Org.ID,

		Email:         params.User.Email,
		EmailVerified: params.User.EmailVerified,
		MFAEnabled:    params.User.MfaEnabled,
		OrgAdmin:      params.UserOrg.IsOrgAdmin,
		UserName:      params.User.Name,
		UserAvatarURL: params.User.AvatarUrl,
		UserOrgRole:   &params.Role.RoleName,
		UserOrgRoleId: &params.Role.ID,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate token pair!"), span, log, counter, opName)
		return nil, err
	}

	newSessionTokenHash, newRefreshTokenHash := tokens.HashTokens(result)

	err = params.Q.CreateSession(ctx, db.CreateSessionParams{
		ID:               result.SessionId,
		UserID:           params.User.ID,
		OrgID:            params.Org.ID,
		SessionTokenHash: newSessionTokenHash,
		RefreshTokenHash: newRefreshTokenHash,
		IpAddress:        ipAddress,
		UserAgent:        userAgent,
		ExpiresAt:        result.RefreshTokenExpiry,
	})

	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to create session in the db!"), span, log, counter, opName)
		return nil, err
	}

	return result, nil
}

type CreateLoginFlowParams struct {
	UserId       string
	Email        string
	Orgs         []db.Org
	FlowReturnTo *string
	MFARequired  bool
}

// CreateLoginFlow creates a new Login flow for the user and returns the flow ID.
func CreateLoginFlow(params CreateLoginFlowParams, ctx context.Context, c *gin.Context, span trace.Span, log *zap.Logger, counter *prometheus.CounterVec, opName string) (string, error) {
	fId, err := uuid.NewV7()
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate flow ID!"), span, log, counter, opName)
		return "", err
	}

	flow := cache.FlowData{
		ID:          fId.String(),
		Type:        cache.FlowTypeLogin,
		UserID:      params.UserId,
		Email:       params.Email,
		Orgs:        params.Orgs,
		MFARequired: params.MFARequired,
		MFAVerified: false,
		ReturnTo:    params.FlowReturnTo,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // Flow expires in 10 minutes
	}
	err = cache.StoreFlow(ctx, flow)

	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to store flow data!"), span, log, counter, opName)
		return "", err
	}
	return fId.String(), nil
}
