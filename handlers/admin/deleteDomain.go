package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

type AdminDeleteDomainHandler struct {
	AdminDeleteDomainCounter *prometheus.CounterVec
}

func NewAdminDeleteDomainHandler() *AdminDeleteDomainHandler {
	return &AdminDeleteDomainHandler{
		AdminDeleteDomainCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "delete_domain",
				Help:      "Total number of AdminDeleteDomain requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminDeleteDomainHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminDeleteDomainCounter)
	router.DELETE("/api/admin/orgs/:orgId/domains/:domain", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminDeleteDomainResponse struct {
	resp.BaseResponse
}

// @Summary AdminDeleteDomain Endpoint
// @Description Handles AdminDeleteDomain requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param domain path string true "Domain"
// @Success 200 {object} AdminDeleteDomainResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId}/domains/{domain} [delete]
func (h *AdminDeleteDomainHandler) Handle(c *gin.Context) {
	h.AdminDeleteDomainCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // Reset inactivity timer for sysadmin users

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminDeleteDomainHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgID, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
		return
	}

	domain := strings.Trim(c.Param("domain"), "/")
	if domain == "" {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
		return
	}
	err = utils.Validator.Var(domain, "required,domain")
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
		return
	}

	domainObj, err := q.RemoveDomainFromOrg(ctx, db.RemoveDomainFromOrgParams{
		OrgID:  orgID,
		Domain: domain,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23503" { // foreign_key_violation
				utils.ProcessError(c, reserr.BadRequest("Failed to delete domain! The domain is in-use by resources!", "The domain is in-use by a resource(s) and cannot be deleted"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
				return
			}
		}
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to delete domain"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to generate audit log ID"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()

		// Log details map
		var details map[string]any = make(map[string]any)
		var actorType db.LogActionActorType
		var adminId *uuid.UUID

		if adminEmail := c.GetString(middlewares.CtxAdminEmail); adminEmail != "" {
			actorType = db.LogActionActorTypeSysadmin
			details["sys.admin.email"] = adminEmail
			details["org.id"] = orgID.String()
		} else if claims, exists := c.Get(middlewares.CtxSessionTokenClaims); exists && claims != nil {
			if nex, ok := claims.(*tokens.NexeresClaims); ok && nex.OrgAdmin {
				actorType = db.LogActionActorTypeOrgadmin
				details["org.admin.email"] = nex.Email
				details["org.id"] = &nex.OrgId
				details["org.name"] = nex.OrgName
				id, err := uuid.Parse(nex.Subject)
				if err != nil {
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
			return
		}

		// Domain details
		details["org.domain.domain"] = domainObj.Domain
		details["org.domain.autoJoin"] = domainObj.AutoJoin
		details["org.domain.autoJoinRoleID"] = domainObj.AutoJoinRoleID
		details["org.domain.autoJoinRoleName"] = domainObj.AutoJoinRoleName
		details["org.domain.verified"] = domainObj.Verified
		details["org.domain.verifiedAt"] = domainObj.VerifiedAt
		details["org.domain.orgId"] = domainObj.OrgID
		details["org.domain.createdAt"] = domainObj.CreatedAt
		details["org.domain.updatedAt"] = domainObj.UpdatedAt

		// Marshal details
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling audit log details"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
			return
		}

		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: actorType,
			LogAction: db.AuditLogActionDelete,
			LogEntity: db.AuditLogEntityOrgDomain,
			Details:   marshalledDetails,
			OrgID:     &orgID,
			UserID:    adminId,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminDeleteDomainCounter, "AdminDeleteDomainHandler.Handle")
		return
	}

	h.AdminDeleteDomainCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminDeleteDomainResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Delete Domain completed successfully.",
		},
	})
}
