package sys_admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type SysAdminGetOrgDetailsHandler struct {
	SysAdminGetOrgDetailsCounter *prometheus.CounterVec
}

func NewSysAdminGetOrgDetailsHandler() *SysAdminGetOrgDetailsHandler {
	return &SysAdminGetOrgDetailsHandler{
		SysAdminGetOrgDetailsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "sys_admin",
				Name:      "get_org_details",
				Help:      "Total number of SysAdminGetOrgDetails requests",
			}, []string{"status"},
		),
	}
}

func (h *SysAdminGetOrgDetailsHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SysAdminGetOrgDetailsCounter)
	router.GET("/api/sys/admin/orgs/:orgId", h.Handle)
}

type SysAdminGetOrgDetailsResponse struct {
	resp.BaseResponse
	Org     db.Org                        `json:"org"`
	Roles   []db.GetMinimalRolesForOrgRow `json:"roles"`
	Domains []db.OrgDomain                `json:"domains"`
}

// @Summary SysAdminGetOrgDetails Endpoint
// @Description Handles SysAdminGetOrgDetails requests
// @Tags sys_admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Success 200 {object} SysAdminGetOrgDetailsResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/sys/admin/orgs/{orgId} [get]
func (h *SysAdminGetOrgDetailsHandler) Handle(c *gin.Context) {
	h.SysAdminGetOrgDetailsCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "SysAdminGetOrgDetailsHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/")) // Validate OrgId
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid org id!", "Invalid uuid provided"), span, log, h.SysAdminGetOrgDetailsCounter, "SysAdminGetOrgDetailsHandler.Handle")
		return
	}

	// We are not using a transaction here, but if we do any writes or we observe
	// inconsistent data then we should use one.

	org, err := store.Querier.GetOrg(ctx, db.GetOrgParams{
		ID: &orgId,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			utils.ProcessError(c, reserr.BadRequest("Organization not found", "No organization found with the provided ID"), span, log, h.SysAdminGetOrgDetailsCounter, "SysAdminGetOrgDetailsHandler.Handle")
			return
		}
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to query organization from the database"), span, log, h.SysAdminGetOrgDetailsCounter, "SysAdminGetOrgDetailsHandler.Handle")
		return
	}

	var roles []db.GetMinimalRolesForOrgRow
	roles, err = store.Querier.GetMinimalRolesForOrg(ctx, orgId)
	if err != nil {
		if err == pgx.ErrNoRows {
			roles = make([]db.GetMinimalRolesForOrgRow, 0) // No roles found, return empty slice
		} else {
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to query roles from the database"), span, log, h.SysAdminGetOrgDetailsCounter, "SysAdminGetOrgDetailsHandler.Handle")
			return
		}
	}

	domains, err := store.Querier.GetOrgDomains(ctx, orgId)
	if err != nil {
		if err == pgx.ErrNoRows {
			domains = make([]db.OrgDomain, 0) // No domains found, return empty slice
		} else {
			utils.ProcessError(c, reserr.InternalServerError(err, "Failed to query organization domains from the database"), span, log, h.SysAdminGetOrgDetailsCounter, "SysAdminGetOrgDetailsHandler.Handle")
			return
		}
	}

	h.SysAdminGetOrgDetailsCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SysAdminGetOrgDetailsResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Get Org Details completed successfully.",
		},
		Org:     org,
		Roles:   roles,
		Domains: domains,
	})
}
