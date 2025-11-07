package sys_admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/prometheus/client_golang/prometheus"
)

type SysAdminGetConfigHandler struct {
	SysAdminGetConfigCounter *prometheus.CounterVec
}

func NewSysAdminGetConfigHandler() *SysAdminGetConfigHandler {
	return &SysAdminGetConfigHandler{
		SysAdminGetConfigCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "sys_admin",
				Name:      "get_config",
				Help:      "Total number of SysAdminGetConfig requests",
			}, []string{"status"},
		),
	}
}

func (h *SysAdminGetConfigHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.SysAdminGetConfigCounter)
	router.GET("/api/sys/admin/config", middlewares.RequireAuth(middlewares.AuthModeSysAdmin), h.Handle)
}

type SysAdminGetConfigResponse struct {
	resp.BaseResponse
	Config *config.CompleteConfig `json:"config"`
}

// @Summary SysAdminGetConfig Endpoint
// @Description Handles SysAdminGetConfig requests
// @Tags sys_admin
// @Accept json
// @Produce json
// @Success 200 {object} SysAdminGetConfigResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/sys/admin/config [get]
func (h *SysAdminGetConfigHandler) Handle(c *gin.Context) {
	h.SysAdminGetConfigCounter.WithLabelValues("total").Inc()
	middlewares.AdminInactivityReset(c) // Reset inactivity timer

	h.SysAdminGetConfigCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &SysAdminGetConfigResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Get Configuration completed successfully.",
		},
		Config: config.C,
	})
}
