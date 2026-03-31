package middlewares

import (
	"net/http"
	"time"

	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/tokens"
)

// AdminInactivityReset resets the inactivity timer for admin sessions.
//
// It retrieves the admin token from the context, and if present, updates the session's
// expiry time based on the configured session timeout. The new expiry time is also set
// in the response header.
//
// If the admin token is not found or if there is an error retrieving or storing the
// session, the expiry time is set to a past time to effectively expire the session.
func AdminInactivityReset(w http.ResponseWriter, r *http.Request) {
	adminToken, ok := r.Context().Value(CtxAdminToken).(string)
	if !ok || adminToken == "" {
		return
	}

	var expiry time.Time
	expiry = time.Now().Add(time.Second * time.Duration(config.C.Admins.SessionTimeoutSeconds))
	session, err := cache.GetAdminSession(r.Context(), adminToken)
	if err != nil {
		expiry = time.Now().Add(time.Hour * (-24)) // set to past time to expire immediately
	} else {
		oldExpiry := session.ExpiresAt
		session.ExpiresAt = expiry
		err = cache.StoreAdminSession(r.Context(), *session)
		if err != nil {
			expiry = oldExpiry // revert to old expiry on error
		}
	}

	// Reset the inactivity timer on each request to an admin endpoint
	w.Header().Set(tokens.AdminTokenExpiryHeaderName, expiry.Format(time.RFC3339))
}
