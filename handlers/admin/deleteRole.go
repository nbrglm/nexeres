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

type AdminDeleteRoleHandler struct {
	AdminDeleteRoleCounter *prometheus.CounterVec
}

func NewAdminDeleteRoleHandler() *AdminDeleteRoleHandler {
	return &AdminDeleteRoleHandler{
		AdminDeleteRoleCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "delete_role",
				Help:      "Total number of AdminDeleteRole requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminDeleteRoleHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminDeleteRoleCounter)
	router.DELETE("/api/admin/orgs/:orgId/roles/:roleId", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminDeleteRoleResponse struct {
	resp.BaseResponse
}

// @Summary AdminDeleteRole Endpoint
// @Description Handles AdminDeleteRole requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} AdminDeleteRoleResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId}/roles/{roleId} [delete]
func (h *AdminDeleteRoleHandler) Handle(c *gin.Context) {
	h.AdminDeleteRoleCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // Reset inactivity timer for admin users

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminDeleteRoleHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
		return
	}

	roleId, err := uuid.Parse(strings.Trim(c.Param("roleId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided role ID is not valid"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
		return
	}

	// Safeguards:
	// The SQL Schema has ON DELETE RESTRICT for roles linked to users, orgs, invites, org_domains, etc.

	role, err := q.DeleteRole(ctx, db.DeleteRoleParams{
		RoleID: roleId,
		OrgID:  orgId,
	})
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			utils.ProcessError(c, reserr.BadRequest("Role not found!", "The specified role does not exist"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
			return
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23503" { // foreign_key_violation
				utils.ProcessError(c, reserr.BadRequest("Failed to delete role! The role is assigned to one or more users, or orgs, or invites or autojoin domains.", "The role is assigned to one or more users and cannot be deleted"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
				return
			}
		}
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to delete role"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to generate audit log ID"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
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
			details["org.id"] = orgId.String()
		} else if claims, exists := c.Get(middlewares.CtxSessionTokenClaims); exists && claims != nil {
			if nex, ok := claims.(*tokens.NexeresClaims); ok && nex.OrgAdmin {
				actorType = db.LogActionActorTypeOrgadmin
				details["org.admin.email"] = nex.Email
				details["org.id"] = &nex.OrgId
				details["org.name"] = nex.OrgName
				id, err := uuid.Parse(nex.Subject)
				if err != nil {
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
			return
		}

		// Role details
		details["org.role.id"] = roleId.String()
		details["org.role.name"] = role.RoleName
		details["org.role.permissions"] = role.Permissions
		details["org.role.description"] = role.RoleDesc
		details["org.role.createdAt"] = role.CreatedAt
		details["org.role.updatedAt"] = role.UpdatedAt

		// Marshal details
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling audit log details"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
			return
		}

		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: actorType,
			LogAction: db.AuditLogActionDelete,
			LogEntity: db.AuditLogEntityRole,
			Details:   marshalledDetails,
			OrgID:     &orgId,
			UserID:    adminId,
			EntityID:  &roleId,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminDeleteRoleCounter, "AdminDeleteRoleHandler.Handle")
		return
	}

	h.AdminDeleteRoleCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminDeleteRoleResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Delete Role completed successfully.",
		},
	})
}
