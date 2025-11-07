package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/handlers/sys_admin"
	"github.com/nbrglm/nexeres/internal/interfaces"
)

func RegisterAPIRoutes(engine *gin.Engine) {
	handlers := []interfaces.Handler{
		NewSignupHandler(),
		NewLoginHandler(),
		NewLogoutHandler(),
		NewRefreshTokenHandler(),
		NewGetFlowHandler(),
		NewSendVerifyEmailHandler(),
	}
	handlers = append(handlers, sys_admin.GetSysAdminHandlers()...)
	for _, handler := range handlers {
		handler.Register(engine)
	}

	// NOTE: For resetting the inactivity timer on admin routes,
	// we need to call `middlewares.AdminInactivityReset()` before any computations
	// in each handler are done.
}
