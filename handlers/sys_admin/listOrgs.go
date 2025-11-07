package sys_admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/nbrglm/nexeres/internal/resquery"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/prometheus/client_golang/prometheus"
)

type SysAdminListOrgsHandler struct {
	SysAdminListOrgsCounter *prometheus.CounterVec
}

func NewSysAdminListOrgsHandler() *SysAdminListOrgsHandler {
	return &SysAdminListOrgsHandler{
		SysAdminListOrgsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "sys_admin",
				Name:      "list_orgs",
				Help:      "Total number of SysAdminListOrgs requests",
			}, []string{"status"},
		),
	}
}

func (h *SysAdminListOrgsHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SysAdminListOrgsCounter)
	router.POST("/api/sys/admin/orgs", middlewares.RequireAuth(middlewares.AuthModeSysAdmin), h.Handle)
}

type SysAdminListOrgsRequest struct {
	Filters    *resquery.QueryFilters    `json:"filters,omitempty" validate:"omitempty"`
	Pagination *resquery.QueryPagination `json:"pagination,omitempty" validate:"omitempty"`
	Sort       []resquery.SortOption     `json:"sort,omitempty" validate:"omitempty,dive,required"`
}

var SysAdminListOrgsQueryFiltersAllowedFields = []string{"name", "slug", "created_at"}
var SysAdminListOrgsQueryFilterAllowedOps = []resquery.QueryFilterOp{
	resquery.OpContains,
	resquery.OpEquals,
	resquery.OpLte,
	resquery.OpGte,
	resquery.OpLt,
	resquery.OpGt,
}
var SysAdminListOrgsQueryFilterAllowedModes = []resquery.FilterMode{
	resquery.FilterModeAND,
	resquery.FilterModeOR,
}
var SysAdminListOrgsQuerySortAllowedFields = []string{"name", "slug", "created_at"}

var SysAdminListOrgsSelectFields = []string{
	"id",
	"name",
	"slug",
	"description",
	"avatar_url",
	"settings",
	"created_at",
	"updated_at",
}

type SysAdminListOrgsResponse struct {
	resp.BaseResponse

	// List of organizations
	Orgs []db.Org `json:"orgs"`

	// Total number of organizations matching the filters (if any)
	Total int `json:"total"`

	// Pagination details, default if not provided in the request
	Pagination *resquery.QueryPagination `json:"pagination"`
}

// @Summary SysAdminListOrgs Endpoint
// @Description Handles SysAdminListOrgs requests
// @Tags sys_admin
// @Accept json
// @Produce json
// @Param request body SysAdminListOrgsRequest true "Request body"
// @Success 200 {object} SysAdminListOrgsResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/sys/admin/orgs [post]
func (h *SysAdminListOrgsHandler) Handle(c *gin.Context) {
	h.SysAdminListOrgsCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "SysAdminListOrgsHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	var req SysAdminListOrgsRequest
	if c.Request.ContentLength < 1 {
		req = SysAdminListOrgsRequest{}
	} else if err := c.ShouldBindJSON(&req); err != nil {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.SysAdminListOrgsCounter, "SysAdminListOrgsHandler.Handle")
		return
	}

	if !req.Filters.Validate(SysAdminListOrgsQueryFiltersAllowedFields, SysAdminListOrgsQueryFilterAllowedOps, SysAdminListOrgsQueryFilterAllowedModes) || !resquery.ValidateSortOptions(req.Sort, SysAdminListOrgsQuerySortAllowedFields) {
		reserr.ProcessError(c, reserr.BadRequest(), span, log, h.SysAdminListOrgsCounter, "SysAdminListOrgsHandler.Handle")
		return
	}

	// Not using the querier or a TX here as we need to build dynamic queries,
	// and it is a single read only.

	filterMode := resquery.FilterModeAND
	if req.Filters != nil {
		filterMode = req.Filters.Mode
	}

	filters := []resquery.FilterOption{}
	if req.Filters != nil {
		filters = req.Filters.Filters
	}

	sortOpts := []resquery.SortOption{}
	if req.Sort != nil {
		sortOpts = req.Sort
	}

	pagination := resquery.DefaultQueryPagination()
	if req.Pagination != nil {
		pagination = req.Pagination
	}

	query, args, err := resquery.GeneratePGSQL("orgs", filterMode, SysAdminListOrgsSelectFields, filters, pagination, sortOpts)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to generate database query"), span, log, h.SysAdminListOrgsCounter, "SysAdminListOrgsHandler.Handle")
		return
	}

	orgs, err := store.PgPool.Query(ctx, query, args...)

	total := 0
	result := []db.Org{}
	for orgs.Next() {
		var org db.Org
		// TODO: optimize total count retrieval, if possible
		err = orgs.Scan(&org.ID, &org.Name, &org.Slug, &org.Description, &org.AvatarUrl, &org.Settings, &org.CreatedAt, &org.UpdatedAt, &total)
		if err != nil {
			reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to parse organization data"), span, log, h.SysAdminListOrgsCounter, "SysAdminListOrgsHandler.Handle")
			return
		}
		result = append(result, org)
	}

	h.SysAdminListOrgsCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SysAdminListOrgsResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation List Orgs completed successfully.",
		},
		Orgs:       result,
		Pagination: pagination,
		Total:      total,
	})
}
