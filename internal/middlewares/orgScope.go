package middlewares

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nbrglm/nexeres/internal/api/contracts"
	"github.com/nbrglm/nexeres/internal/tokens"
)

// RequireOrgScope is a middleware that ensures the request is scoped to the organization
// specified in the URL parameter "orgId". If the parameter is absent, the middleware
// allows the request to proceed without org scope validation.
//
// This middleware MUST BE ATTACHED AFTER PopulateAuthContext to ensure that
// authentication information is available in the context.
func RequireOrgScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgId := chi.URLParam(r, "orgId")
		if orgId == "" {
			// No orgId parameter present; skip org scope check
			next.ServeHTTP(w, r)
			return
		}

		claimsObj := r.Context().Value(CtxSessionTokenClaims)
		_, sysAdminExists := r.Context().Value(CtxAdminEmail).(string)
		if claimsObj == nil && !sysAdminExists {
			contracts.Unauthorized("Unauthorized: missing session token or admin token", "Please add either a session token or an admin token!").Write(w)
			return
		}

		if sysAdminExists {
			// Sysadmins have access to all orgs
			next.ServeHTTP(w, r)
			return
		}

		if claims, ok := claimsObj.(*tokens.NexeresClaims); ok {
			if claims.OrgId.String() != orgId {
				contracts.Unauthorized("Unauthorized", "The session token does not have access to the requested organization").Write(w)
				return
			}
		} else {
			contracts.Unauthorized("Unauthorized: invalid session token", "Please provide a valid session token!").Write(w)
			return
		}

		// Org ID in session token matches the requested org ID
		next.ServeHTTP(w, r)
	})
}
