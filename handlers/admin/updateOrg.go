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
	"github.com/nbrglm/nexeres/internal/models"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type AdminUpdateOrgHandler struct {
	AdminUpdateOrgCounter *prometheus.CounterVec
}

func NewAdminUpdateOrgHandler() *AdminUpdateOrgHandler {
	return &AdminUpdateOrgHandler{
		AdminUpdateOrgCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "update_org",
				Help:      "Total number of AdminUpdateOrg requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminUpdateOrgHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminUpdateOrgCounter)
	router.POST("/api/admin/orgs/:orgId", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminUpdateOrgRequest struct {
	Name        string             `json:"name" validate:"required"`
	Description *string            `json:"description,omitempty"`
	Settings    models.OrgSettings `json:"settings" validate:"required"`
}

type AdminUpdateOrgResponse struct {
	resp.BaseResponse
}

// @Summary AdminUpdateOrg Endpoint
// @Description Handles AdminUpdateOrg requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param request body AdminUpdateOrgRequest true "Request body"
// @Success 200 {object} AdminUpdateOrgResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId} [post]
func (h *AdminUpdateOrgHandler) Handle(c *gin.Context) {
	h.AdminUpdateOrgCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // Reset inactivity timer for admin users

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminUpdateOrgHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req AdminUpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
		return
	}

	org, err := q.UpdateOrg(ctx, db.UpdateOrgParams{
		ID:          &orgId,
		Name:        &req.Name,
		Description: req.Description,
		Settings:    req.Settings,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to update organization"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log ID"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()
		// Log details map
		var details map[string]any = make(map[string]any)
		var actorType db.LogActionActorType
		var adminId *uuid.UUID

		// Populate details
		details["org.name"] = org.Name
		if org.Description != nil {
			details["org.description"] = *org.Description
		}
		details["org.id"] = &orgId
		details["org.settings"] = org.Settings
		details["org.avatarUrl"] = org.AvatarUrl
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
					utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the org admin ID for audit log"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
					return
				}
				adminId = &id
				details["org.admin.id"] = adminId
			}
		} else {
			utils.ProcessError(c, reserr.InternalServerError(fmt.Errorf("could not determine actor type for audit log"), "Could not determine actor type for audit log"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
			return
		}
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling the audit log details"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
			return
		}
		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: actorType,
			LogAction: db.AuditLogActionUpdate,
			LogEntity: db.AuditLogEntityOrg,
			Details:   marshalledDetails,
			OrgID:     &orgId,
			UserID:    adminId,
			EntityID:  &orgId,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminUpdateOrgCounter, "AdminUpdateOrgHandler.Handle")
		return
	}

	h.AdminUpdateOrgCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminUpdateOrgResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Update Org completed successfully.",
		},
	})
}
