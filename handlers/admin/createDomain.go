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

type AdminCreateDomainHandler struct {
	AdminCreateDomainCounter *prometheus.CounterVec
}

func NewAdminCreateDomainHandler() *AdminCreateDomainHandler {
	return &AdminCreateDomainHandler{
		AdminCreateDomainCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "create_domain",
				Help:      "Total number of AdminCreateDomain requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminCreateDomainHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminCreateDomainCounter)
	router.PUT("/api/admin/orgs/:orgId/domains", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminCreateDomainRequest struct {
	Domain           string     `json:"domain" binding:"required,domain"`
	AutoJoin         *bool      `json:"autoJoin,omitempty"`
	AutoJoinRoleId   *uuid.UUID `json:"autoJoinRoleId,omitempty" binding:"omitempty,uuidv7"`
	AutoJoinRoleName *string    `json:"autoJoinRoleName,omitempty"`
}

type AdminCreateDomainResponse struct {
	resp.BaseResponse
	Domain db.OrgDomain `json:"domain"`
}

// @Summary AdminCreateDomain Endpoint
// @Description Handles AdminCreateDomain requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param request body AdminCreateDomainRequest true "Request body"
// @Success 200 {object} AdminCreateDomainResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId}/domains [put]
func (h *AdminCreateDomainHandler) Handle(c *gin.Context) {
	h.AdminCreateDomainCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // reset inactivity timer

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminCreateDomainHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req AdminCreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
		return
	}

	domain, err := q.CreateOrgDomain(ctx, db.CreateOrgDomainParams{
		OrgID:            orgId,
		Domain:           req.Domain,
		AutoJoin:         req.AutoJoin,
		AutoJoinRoleID:   req.AutoJoinRoleId,
		AutoJoinRoleName: req.AutoJoinRoleName,
	})

	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr != nil && pgErr.Code == "23505" {
		// Get the organization to which the domain was added
		org, err := q.GetOrgByDomain(ctx, db.GetOrgByDomainParams{
			Domain: req.Domain,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the domain, couldn't fetch which org the domain is registered to!"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
			return
		}
		// Happens when there is a conflict with an existing domain
		utils.ProcessError(c, reserr.BadRequest(fmt.Sprintf("This domain is already claimed by %s organization!", org.Org.Name), "The domain already exists in another organization"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
		return
	}

	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the domain"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log ID"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()
		// Log details map
		var details map[string]any = make(map[string]any)
		var actorType db.LogActionActorType
		var adminId *uuid.UUID

		// Populate details
		details["org.domain.name"] = domain.Domain
		if adminEmail := c.GetString(middlewares.CtxAdminEmail); adminEmail != "" {
			actorType = db.LogActionActorTypeSysadmin
			details["sys.admin.email"] = adminEmail
		} else if claims, exists := c.Get(middlewares.CtxSessionTokenClaims); exists && claims != nil {
			if nex, ok := claims.(*tokens.NexeresClaims); ok && nex.OrgAdmin {
				actorType = db.LogActionActorTypeOrgadmin
				details["org.admin.email"] = nex.Email
				details["org.id"] = &nex.OrgId
				details["org.name"] = nex.OrgName
				id, err := uuid.Parse(nex.Subject)
				if err != nil {
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
			return
		}
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling the audit log details"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
			return
		}
		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: actorType,
			LogAction: db.AuditLogActionCreate,
			LogEntity: db.AuditLogEntityOrgDomain,
			Details:   marshalledDetails,
			OrgID:     &orgId,
			UserID:    adminId,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminCreateDomainCounter, "AdminCreateDomainHandler.Handle")
		return
	}

	h.AdminCreateDomainCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminCreateDomainResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Create Domain completed successfully.",
		},
		Domain: domain,
	})
}
