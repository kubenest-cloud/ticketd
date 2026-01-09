package web

import (
	"log/slog"
	"net/http"
)

// basicAuth is a middleware that protects routes with HTTP Basic Authentication.
// It checks the provided credentials against the configured admin username and password.
// Returns 401 Unauthorized if credentials are missing or invalid.
//
// If DisableAuth is set to true in the configuration, authentication is bypassed entirely.
// This is useful when deploying behind external authentication proxies like oauth2-proxy,
// Authelia, or similar solutions.
//
// SECURITY WARNING: Only disable authentication when using a trusted external auth proxy.
// Never expose TicketD directly to the internet with authentication disabled.
func (a *App) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if disabled (for use with external auth proxies)
		if a.Cfg.DisableAuth {
			slog.Debug("Authentication bypassed (external auth mode)", "path", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Perform standard HTTP Basic Auth
		user, pass, ok := r.BasicAuth()
		if !ok || user != a.Cfg.AdminUser || pass != a.Cfg.AdminPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="TicketD"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
