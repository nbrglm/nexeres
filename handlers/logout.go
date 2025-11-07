package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type LogoutHandler struct {
	LogoutCounter *prometheus.CounterVec
}

func NewLogoutHandler() *LogoutHandler {
	return &LogoutHandler{
		LogoutCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nexeres",
				Subsystem: "auth",
				Name:      "logout",
				Help:      "Total number of Logout requests",
			}, []string{"status"},
		),
	}
}

func (h *LogoutHandler) Register(router *gin.Engine) {
	metrics.RegisterCollector(h.LogoutCounter)
	router.POST("/api/auth/logout", h.Handle)
}

type LogoutResponse struct {
	resp.BaseResponse
	// TODO: Define the response structure
}

// @Summary Logout Endpoint
// @Description Handles Logout requests
// @Tags auth
// @Accept json
// @Produce json
// @Param X-NEXERES-Session-Token header string false "Session token"
// @Param X-NEXERES-Refresh-Token header string false "Refresh token"
// @Success 200 {object} LogoutResponse
// @Failure 400 {object} reserr.ErrorResponse
// @Failure 401 {object} reserr.ErrorResponse
// @Failure 500 {object} reserr.ErrorResponse
// @Router /api/auth/logout [post]
func (h *LogoutHandler) Handle(c *gin.Context) {
	h.LogoutCounter.WithLabelValues("total").Inc()

	ctx, log, span := obs.WithContext(c.Request.Context(), "LogoutHandler.Handle")
	defer span.End() // Ensure the span is ended when the function returns to prevent memory leaks

	tx, err := store.PgPool.Begin(ctx)
	if err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to begin transaction"), span, log, h.LogoutCounter, "LogoutHandler.Handle")
		return
	}
	defer tx.Rollback(ctx)
	q := store.Querier.WithTx(tx)

	sessionToken := c.GetString(middlewares.CtxSessionToken)
	refreshToken := c.GetString(middlewares.CtxRefreshToken)

	if sessionToken != "" {
		claims := c.MustGet(middlewares.CtxSessionTokenClaims).(*tokens.NexeresClaims)
		var id uuid.UUID
		id, err = uuid.Parse(claims.ID)
		if err == nil {
			err = q.DeleteSession(ctx, db.DeleteSessionParams{
				ID: &id,
			})
		}
	} else {
		_, tokenHash := tokens.HashTokens(&tokens.Tokens{
			RefreshToken: refreshToken,
		})
		err = q.DeleteSession(ctx, db.DeleteSessionParams{
			RefreshTokenHash: &tokenHash,
		})
	}

	if err != nil {
		utils.ProcessError(c, reserr.InternalServerError(err, "Failed to revoke session"), span, log, h.LogoutCounter, "LogoutHandler.Handle")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		reserr.ProcessError(c, reserr.InternalServerError(err, "Failed to commit transaction"), span, log, h.LogoutCounter, "LogoutHandler.Handle")
		return
	}

	h.LogoutCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, &LogoutResponse{
		BaseResponse: resp.BaseResponse{
			Success: true,
			Message: "Operation Logout completed successfully.",
		},
	})
}
