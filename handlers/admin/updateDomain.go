package admin

import (
	"encoding/json"
	"fmt"
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

type AdminUpdateDomainHandler struct {
	AdminUpdateDomainCounter *prometheus.CounterVec
}

func NewAdminUpdateDomainHandler() *AdminUpdateDomainHandler {
	return &AdminUpdateDomainHandler{
		AdminUpdateDomainCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "update_domain",
				Help:      "Total number of AdminUpdateDomain requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminUpdateDomainHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminUpdateDomainCounter)
	router.POST("/api/admin/orgs/:orgId/domains/:domain", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminUpdateDomainRequest struct {
	AutoJoin         *bool      `json:"autoJoin,omitempty" validate:"omitempty"`
	AutoJoinRoleId   *uuid.UUID `json:"autoJoinRoleId,omitempty" validate:"omitempty,uuidv7"`
	AutoJoinRoleName *string    `json:"autoJoinRoleName,omitempty" validate:"required_with=AutoJoinRoleId"`
}

type AdminUpdateDomainResponse struct {
	resp.BaseResponse
	Domain db.OrgDomain `json:"domain"`
}

// @Summary AdminUpdateDomain Endpoint
// @Description Handles AdminUpdateDomain requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param domain path string true "Domain"
// @Param request body AdminUpdateDomainRequest true "Request body"
// @Success 200 {object} AdminUpdateDomainResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/:orgId/domains/:domain [post]
func (h *AdminUpdateDomainHandler) Handle(c *gin.Context) {
	h.AdminUpdateDomainCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // Reset inactivity timer for admin users

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminUpdateDomainHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req AdminUpdateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}

	domain := strings.Trim(c.Param("domain"), "/")
	if domain == "" {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}
	err = utils.Validator.Var(domain, "required,domain")
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}

	updatedDomain, err := q.UpdateOrgDomain(ctx, db.UpdateOrgDomainParams{
		OrgID:            orgId,
		Domain:           domain,
		AutoJoin:         req.AutoJoin,
		AutoJoinRoleID:   req.AutoJoinRoleId,
		AutoJoinRoleName: req.AutoJoinRoleName,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to update domain"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log ID"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()
		// Log details map
		var details map[string]any = make(map[string]any)
		var actorType db.LogActionActorType
		var adminId *uuid.UUID

		// Populate details
		details["org.domain"] = updatedDomain.Domain
		details["org.domain.autoJoin.role.name"] = updatedDomain.AutoJoinRoleName
		details["org.domain.autoJoin.role.id"] = updatedDomain.AutoJoinRoleID
		details["org.domain.autoJoin"] = updatedDomain.AutoJoin

		if adminEmail := c.GetString(middlewares.CtxAdminEmail); adminEmail != "" {
			actorType = db.LogActionActorTypeSysadmin
			details["org.id"] = updatedDomain.OrgID
			details["sys.admin.email"] = adminEmail
		} else if claims, exists := c.Get(middlewares.CtxSessionTokenClaims); exists && claims != nil {
			if nex, ok := claims.(*tokens.NexeresClaims); ok && nex.OrgAdmin {
				actorType = db.LogActionActorTypeOrgadmin
				details["org.admin.email"] = nex.Email
				details["org.id"] = &nex.OrgId
				details["org.name"] = nex.OrgName
				id, err := uuid.Parse(nex.Subject)
				if err != nil {
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
			return
		}
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling the audit log details"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
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
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminUpdateDomainCounter, "AdminUpdateDomainHandler.Handle")
		return
	}

	h.AdminUpdateDomainCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminUpdateDomainResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Update Domain completed successfully.",
		},
		Domain: updatedDomain,
	})
}
