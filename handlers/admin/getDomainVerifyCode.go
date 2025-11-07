package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type AdminGetDomainVerifyCodeHandler struct {
	AdminGetDomainVerifyCodeCounter *prometheus.CounterVec
}

func NewAdminGetDomainVerifyCodeHandler() *AdminGetDomainVerifyCodeHandler {
	return &AdminGetDomainVerifyCodeHandler{
		AdminGetDomainVerifyCodeCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "admin",
				Name:      "get_domain_verify_code",
				Help:      "Total number of AdminGetDomainVerifyCode requests",
			}, []string{"status"},
		),
	}
}

func (h *AdminGetDomainVerifyCodeHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.AdminGetDomainVerifyCodeCounter)
	router.GET("/api/admin/orgs/:orgId/domains/:domain/verify/code", middlewares.RequireAuth(middlewares.AuthModeAnyAdmin), h.Handle)
}

type AdminGetDomainVerifyCodeResponse struct {
	resp.BaseResponse
	Code string `json:"code"`
}

// @Summary AdminGetDomainVerifyCode Endpoint
// @Description Handles AdminGetDomainVerifyCode requests
// @Tags admin
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param domain path string true "Domain"
// @Success 200 {object} AdminGetDomainVerifyCodeResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/admin/orgs/{orgId}/domains/{domain}/verify/code [get]
func (h *AdminGetDomainVerifyCodeHandler) Handle(c *gin.Context) {
	h.AdminGetDomainVerifyCodeCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "AdminGetDomainVerifyCodeHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.AdminGetDomainVerifyCodeCounter, "AdminGetDomainVerifyCodeHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	orgId, err := uuid.Parse(strings.Trim(c.Param("orgId"), "/"))
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided organization ID is not valid"), span, log, h.AdminGetDomainVerifyCodeCounter, "AdminGetDomainVerifyCodeHandler.Handle")
		return
	}

	domain := strings.Trim(c.Param("domain"), "/")
	if domain == "" {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminGetDomainVerifyCodeCounter, "AdminGetDomainVerifyCodeHandler.Handle")
		return
	}
	err = utils.Validator.Var(domain, "required,domain")
	if err != nil {
		utils.ProcessError(c, reserr.BadRequest("Invalid request!", "The provided domain is not valid"), span, log, h.AdminGetDomainVerifyCodeCounter, "AdminGetDomainVerifyCodeHandler.Handle")
		return
	}

	domain = fmt.Sprintf("_nexeres-domain-verification.%s", domain)

	org, err := q.GetOrgById(ctx, orgId) // Get the org details
	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to get organization details"), span, log, h.AdminGetDomainVerifyCodeCounter, "AdminGetDomainVerifyCodeHandler.Handle")
		return
	}

	code := tokens.GenerateDomainVerifyCode(domain, orgId.String(), org.DomainVerificationSecret)

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.AdminGetDomainVerifyCodeCounter, "AdminGetDomainVerifyCodeHandler.Handle")
		return
	}

	h.AdminGetDomainVerifyCodeCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &AdminGetDomainVerifyCodeResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Get Domain Verification Code completed successfully.",
		},
		Code: code,
	})
}
