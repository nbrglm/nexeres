package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/obs"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/resp"
	"github.com/prometheus/client_golang/prometheus"
)

type GetFlowHandler struct {
	GetFlowCounter *prometheus.CounterVec
}

func NewGetFlowHandler() *GetFlowHandler {
	return &GetFlowHandler{
		GetFlowCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "get_flow",
				Help:      "Total number of GetFlow requests",
			}, []string{"status"},
		),
	}
}

func (h *GetFlowHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.GetFlowCounter)
	router.GET("/api/auth/flow/:flowId", h.Handle)
}

type GetFlowResponse struct {
	resp.BaseResponse
	Flow cache.FlowData `json:"flow"`
}

// @Summary GetFlow Endpoint
// @Description Handles GetFlow requests
// @Tags auth
// @Accept json
// @Produce json
// @Param flowId path string true "Flow ID"
// @Success 200 {object} GetFlowResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/flow/{flowId} [get]
func (h *GetFlowHandler) Handle(c *gin.Context) {
	h.GetFlowCounter.WithLabelValues("total").Inc()

	_, _, span := obs.WithContext(c.Request.Context(), "GetFlowHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	flowId := strings.TrimSuffix(strings.TrimSpace(c.Param("flowId")), "/")
	flow, err := cache.GetFlow(c.Request.Context(), flowId)

	if err != nil {
		if err == cache.ErrKeyNotFound {
			h.GetFlowCounter.WithLabelValues("not_found").Inc()
			c.JSON(http.StatusBadRequest, reserr.BadRequest("Flow not found", "Not Found"))
			return
		}
		h.GetFlowCounter.WithLabelValues("error").Inc()
		c.JSON(http.StatusBadRequest, reserr.BadRequest("Error fetching flow data", "Bad Request"))
		return
	}

	h.GetFlowCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &GetFlowResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation GetFlow completed successfully.",
		},
		Flow: *flow,
	})
}
