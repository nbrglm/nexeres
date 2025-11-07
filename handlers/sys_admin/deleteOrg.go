package sys_admin

import (
	"encoding/json"
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
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type SysAdminDeleteOrgHandler struct {
	SysAdminDeleteOrgCounter *prometheus.CounterVec
}

func NewSysAdminDeleteOrgHandler() *SysAdminDeleteOrgHandler {
	return &SysAdminDeleteOrgHandler{
		SysAdminDeleteOrgCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "sys_admin",
				Name:      "delete_org",
				Help:      "Total number of SysAdminDeleteOrg requests",
			}, []string{"status"},
		),
	}
}

func (h *SysAdminDeleteOrgHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SysAdminDeleteOrgCounter)
	router.DELETE("/api/sys/admin/orgs/:orgId", middlewares.RequireAuth(middlewares.AuthModeSysAdmin), h.Handle)
}

type SysAdminDeleteOrgResponse struct {
	resp.BaseResponse
}

// @Summary SysAdminDeleteOrg Endpoint
// @Description Handles SysAdminDeleteOrg requests
// @Tags sys_admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Success 200 {object} SysAdminDeleteOrgResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/sys/admin/orgs/{orgId} [delete]
func (h *SysAdminDeleteOrgHandler) Handle(c *gin.Context) {
	h.SysAdminDeleteOrgCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // reset inactivity timer

	ctx, log, span := obs.WithContext(c.Request.Context(), "SysAdminDeleteOrgHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgID, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
		return
	}

	// Safeguards:
	// The SQL Schema has ON DELETE RESTRICT for all foreign keys pointing to orgs,
	// so if there are any dependent records (users, devices, etc.), the deletion will fail.
	// This is to prevent accidental data loss.

	// Thus, only if there are existing sessions (possibly due to cleanup delays), this deletion will proceed.
	// Otherwise, it will be blocked by the foreign key constraints.
	deletedOrg, err := q.DeleteOrgById(ctx, orgID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23503" { // foreign_key_violation
				utils.ProcessError(c, reserr.BadRequest("Failed to delete organization! Either custom roles, pending invitations, email domains or users exists for this organization.", "The organization has dependent records (e.g., users, etc.) and cannot be deleted"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
				return
			}
		}
		utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while trying to delete the organization"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to generate audit log ID"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()

		var details map[string]any = make(map[string]any)
		details["org.id"] = orgID.String()
		details["org.name"] = deletedOrg.Name
		details["org.slug"] = deletedOrg.Slug
		details["org.avatarUrl"] = deletedOrg.AvatarUrl
		details["org.createdAt"] = deletedOrg.CreatedAt
		details["org.description"] = deletedOrg.Description
		details["org.settings"] = deletedOrg.Settings
		details["sys.admin.email"] = c.GetString(middlewares.CtxAdminEmail)
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling audit log details"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
			return
		}

		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: db.LogActionActorTypeSysadmin,
			LogAction: db.AuditLogActionDelete,
			LogEntity: db.AuditLogEntityOrg,
			Details:   marshalledDetails,
			OrgID:     &orgID,
			EntityID:  &orgID,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.SysAdminDeleteOrgCounter, "SysAdminDeleteOrgHandler.Handle")
		return
	}

	h.SysAdminDeleteOrgCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SysAdminDeleteOrgResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Delete Organization completed successfully.",
		},
	})
}
