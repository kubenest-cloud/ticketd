package web

import (
	"net/http"
)

// basicAuth is a middleware that protects routes with HTTP Basic Authentication.
// It checks the provided credentials against the configured admin username and password.
// Returns 401 Unauthorized if credentials are missing or invalid.
func (a *App) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != a.Cfg.AdminUser || pass != a.Cfg.AdminPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="TicketD"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
