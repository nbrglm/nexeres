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

type AdminCreateRoleHandler struct {
	AdminCreateRoleCounter *prometheus.CounterVec
}

func NewAdminCreateRoleHandler() *AdminCreateRoleHandler {
	return &AdminCreateRoleHandler{
		AdminCreateRoleCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "create_role",
				Help:      "Total number of CreateRole requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminCreateRoleHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminCreateRoleCounter)
	router.PUT("/api/admin/orgs/:orgId/roles", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminCreateRoleRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description *string  `json:"description" validate:"omitempty"`
	Permissions []string `json:"permissions,omitempty" validate:"omitempty,dive,required"`
}

type AdminCreateRoleResponse struct {
	resp.BaseResponse
	Role db.Role `json:"role"`
}

// @Summary CreateRole Endpoint
// @Description Handles CreateRole requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param request body AdminCreateRoleRequest true "Request body"
// @Success 200 {object} AdminCreateRoleResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId}/roles [put]
func (h *AdminCreateRoleHandler) Handle(c *gin.Context) {
	h.AdminCreateRoleCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // reset inactivity timer

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminCreateRoleHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req AdminCreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate role ID"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
		return
	}

	permissions := req.Permissions
	if permissions == nil {
		permissions = make([]string, 0)
	}

	role, err := q.CreateRole(ctx, db.CreateRoleParams{
		ID:          id,
		OrgID:       orgId,
		RoleName:    req.Name,
		RoleDesc:    req.Description,
		Permissions: permissions,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to create role"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log ID"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()
		// Log details map
		var details map[string]any = make(map[string]any)
		var actorType db.LogActionActorType
		var adminId *uuid.UUID

		// Populate details
		details["org.role.name"] = role.RoleName
		details["org.role.id"] = role.ID
		details["org.role.permissions"] = role.Permissions
		if role.RoleDesc != nil {
			details["org.role.description"] = role.RoleDesc
		}
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
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
			return
		}
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling the audit log details"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
			return
		}
		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: actorType,
			LogAction: db.AuditLogActionCreate,
			LogEntity: db.AuditLogEntityRole,
			Details:   marshalledDetails,
			OrgID:     &orgId,
			UserID:    adminId,
			EntityID:  &role.ID,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminCreateRoleCounter, "AdminCreateRoleHandler.Handle")
		return
	}

	h.AdminCreateRoleCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminCreateRoleResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Create Role completed successfully.",
		},
	})
}
