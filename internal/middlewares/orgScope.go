package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/internal/reserr"
	"github.com/nbrglm/nexeres/internal/tokens"
)

// RequireOrgScope is a middleware that ensures the request is scoped to the organization
// specified in the URL parameter "orgId". If the parameter is absent, the middleware
// allows the request to proceed without org scope validation.
//
// This middleware MUST BE ATTACHED AFTER PopulateAuthContext to ensure that
// authentication information is available in the context.
func RequireOrgScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgId := c.Param("orgId")
		if orgId == "" {
			// No orgId parameter present; skip org scope check
			c.Next()
			return
		}

		claimsObj, sessionExists := c.Get(CtxSessionTokenClaims)
		_, sysAdminExists := c.Get(CtxAdminEmail)
		if !sessionExists && !sysAdminExists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, reserr.Unauthorized("Unauthorized: missing session token or admin token", "Please add either a session token or an admin token!").Filter())
			return
		}

		if sysAdminExists {
			// Sysadmins have access to all orgs
			c.Next()
			return
		}

		if claims, ok := claimsObj.(*tokens.NexeresClaims); ok {
			if claims.OrgId.String() != orgId {
				c.AbortWithStatusJSON(http.StatusUnauthorized, reserr.Unauthorized("Unauthorized", "The session token does not have access to the requested organization").Filter())
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, reserr.Unauthorized("Unauthorized: invalid session token", "Please provide a valid session token!").Filter())
			return
		}

		// Org ID in session token matches the requested org ID
		c.Next()
	}
}
