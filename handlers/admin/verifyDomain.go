package admin

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nbrglm/nexeres/config"
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
)

type AdminVerifyDomainHandler struct {
	AdminVerifyDomainCounter *prometheus.CounterVec
}

func NewAdminVerifyDomainHandler() *AdminVerifyDomainHandler {
	return &AdminVerifyDomainHandler{
		AdminVerifyDomainCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "verify_domain",
				Help:      "Total number of AdminVerifyDomain requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminVerifyDomainHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminVerifyDomainCounter)
	router.POST("/api/admin/orgs/:orgId/domains/:domain/verify", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminVerifyDomainResponse struct {
	resp.BaseResponse
	Verified bool `json:"verified"`
}

// @Summary AdminVerifyDomain Endpoint
// @Description Handles AdminVerifyDomain requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param domain path string true "Domain to verify"
// @Success 200 {object} AdminVerifyDomainResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId}/domains/{domain}/verify [post]
func (h *AdminVerifyDomainHandler) Handle(c *gin.Context) {
	h.AdminVerifyDomainCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminVerifyDomainHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	domain := strings.Trim(c.Param("domain"), "/")
	if domain == "" {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}
	err = utils.Validator.Var(domain, "required,domain")
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	domain = fmt.Sprintf("_nexeres-domain-verification.%s", domain)

	// Lookup TXT records
	txtRecords, err := net.DefaultResolver.LookupTXT(ctx, domain)
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "Failed to lookup TXT records"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	// Check if the domain is verified
	if len(txtRecords) == 0 || (len(txtRecords) == 1 && txtRecords[0] == "") || len(txtRecords) > 1 {
		utils.ProcessError(c, reserr.BadRequest("Invalid request! No valid TXT records found for domain verification", "No valid TXT records found for domain verification"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	org, err := q.GetOrgById(ctx, orgId) // Get the org details
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to get organization details"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	if !tokens.ValidateDomainVerifyCode(domain, orgId.String(), org.DomainVerificationSecret, txtRecords[0]) {
		utils.ProcessError(c, reserr.BadRequest("Invalid request! Domain verification record is invalid", "Domain verification record is invalid"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	// Update domain verification status
	err = store.Querier.UpdateOrgDomainVerification(ctx, db.UpdateOrgDomainVerificationParams{
		Verified: true,
		Domain:   strings.TrimPrefix(domain, "_nexeres-domain-verification."),
		OrgID:    orgId,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to verify domain!"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log ID"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()
		// Log details map
		var details map[string]any = make(map[string]any)
		var actorType db.LogActionActorType
		var adminId *uuid.UUID

		// Populate details
		details["org.domain"] = domain
		details["org.domain.verified"] = true

		if adminEmail := c.GetString(middlewares.CtxAdminEmail); adminEmail != "" {
			actorType = db.LogActionActorTypeSysadmin
			details["org.id"] = orgId
			details["sys.admin.email"] = adminEmail
		} else if claims, exists := c.Get(middlewares.CtxSessionTokenClaims); exists && claims != nil {
			if nex, ok := claims.(*tokens.NexeresClaims); ok && nex.OrgAdmin {
				actorType = db.LogActionActorTypeOrgadmin
				details["org.admin.email"] = nex.Email
				details["org.id"] = &nex.OrgId
				details["org.name"] = nex.OrgName
				id, err := uuid.Parse(nex.Subject)
				if err != nil {
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
			return
		}
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling the audit log details"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
			return
		}
		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: actorType,
			LogAction: db.AuditLogActionUpdate,
			LogEntity: db.AuditLogEntityOrgDomain,
			Details:   marshalledDetails,
			OrgID:     &orgId,
			UserID:    adminId,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminVerifyDomainCounter, "AdminVerifyDomainHandler.Handle")
		return
	}

	h.AdminVerifyDomainCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminVerifyDomainResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation AdminVerifyDomain completed successfully.",
		},
		Verified: true,
	})
}
