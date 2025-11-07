package sys_admin

import (
	"encoding/json"
	"net/http"
	"net/netip"

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
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type SysAdminCreateOrgHandler struct {
	SysAdminCreateOrgCounter *prometheus.CounterVec
}

func NewSysAdminCreateOrgHandler() *SysAdminCreateOrgHandler {
	return &SysAdminCreateOrgHandler{
		SysAdminCreateOrgCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "sys_admin",
				Name:      "create_org",
				Help:      "Total number of SysAdminCreateOrg requests",
			}, []string{"status"},
		),
	}
}

func (h *SysAdminCreateOrgHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SysAdminCreateOrgCounter)
	router.PUT("/api/sys/admin/orgs", middlewares.RequireAuth(middlewares.AuthModeSysAdmin), h.Handle)
}

type SysAdminCreateOrgRequest struct {
	Name        string             `json:"name" validate:"required"`
	Slug        string             `json:"slug" validate:"required,urlslug"`
	Description *string            `json:"description" validate:"omitempty"`
	AvatarURL   *string            `json:"avatarURL,omitempty" validate:"omitempty,url"`
	Settings    models.OrgSettings `json:"settings" validate:"required"`
}

type SysAdminCreateOrgResponse struct {
	resp.BaseResponse
	OrgId string `json:"orgId"`
}

// @Summary SysAdminCreateOrg Endpoint
// @Description Handles SysAdminCreateOrg requests
// @Tags sys_admin
// @Accept json
// @Produce json
// @Param request body SysAdminCreateOrgRequest true "Request body"
// @Success 200 {object} SysAdminCreateOrgResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/sys/admin/orgs [put]
func (h *SysAdminCreateOrgHandler) Handle(c *gin.Context) {
	h.SysAdminCreateOrgCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // reset inactivity timer

	ctx, log, span := obs.WithContext(c.Request.Context(), "SysAdminCreateOrgHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req SysAdminCreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
		return
	}

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	id, err := uuid.NewV7()
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to generate organization ID"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
		return
	}

	err = q.CreateOrg(ctx, db.CreateOrgParams{
		ID:          id,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		AvatarUrl:   req.AvatarURL,
		Settings:    req.Settings,
	})
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to create organization in database"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
		return
	}

	if config.C.Security.AuditLogs.Enable {
		logId, err := uuid.NewV7()
		if err != nil {
			reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to generate audit log ID"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
			return
		}
		ipAddr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while parsing the client IP address"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
			return
		}
		userAgent := c.Request.UserAgent()

		var details map[string]any = make(map[string]any)
		details["org.id"] = id.String()
		details["org.name"] = req.Name
		details["org.slug"] = req.Slug
		details["sys.admin.email"] = c.GetString(middlewares.CtxAdminEmail)
		marshalledDetails, err := json.Marshal(details)
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while marshalling audit log details"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
			return
		}

		err = q.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:        logId,
			ActorType: db.LogActionActorTypeSysadmin,
			LogAction: db.AuditLogActionCreate,
			LogEntity: db.AuditLogEntityOrg,
			Details:   marshalledDetails,
			OrgID:     &id,
			EntityID:  &id,
			IpAddress: &ipAddr,
			UserAgent: &userAgent,
		})
		if err != nil {
			utils.ProcessError(c, reserr.InternalServerError(err, "An error occurred while creating the audit log"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.SysAdminCreateOrgCounter, "SysAdminCreateOrgHandler.Handle")
		return
	}

	h.SysAdminCreateOrgCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SysAdminCreateOrgResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Create Organization completed successfully.",
		},
		OrgId: id.String(),
	})
}
