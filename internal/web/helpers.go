package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// publicBaseURL returns the base URL for public-facing endpoints.
// If TICKETD_PUBLIC_BASE_URL is configured, it uses that.
// Otherwise, it infers the URL from the request (scheme + host).
func (a *App) publicBaseURL(r *http.Request) string {
	if a.Cfg.PublicBaseURL != "" {
		return strings.TrimRight(a.Cfg.PublicBaseURL, "/")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// baseURLForAdmin returns the base URL and an optional warning note for admin display.
// The warning note is shown when the public base URL is not configured,
// as embed links may be unstable without it.
func (a *App) baseURLForAdmin(r *http.Request) (string, string) {
	if a.Cfg.PublicBaseURL != "" {
		return strings.TrimRight(a.Cfg.PublicBaseURL, "/"), ""
	}
	return a.publicBaseURL(r), "Set TICKETD_PUBLIC_BASE_URL in production for stable embed links."
}

// debugEnabled checks if debug logging is enabled via the TICKETD_DEBUG environment variable.
// Set TICKETD_DEBUG=1 to enable verbose logging of CORS and submission details.
func debugEnabled() bool {
	return os.Getenv("TICKETD_DEBUG") == "1"
}

// parseID parses a URL parameter as an int64 ID.
// Returns an error if the value is not a valid integer.
func parseID(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

// parsePage extracts the page number from the query string.
// Defaults to page 1 if not specified or invalid.
// Only positive integers are accepted.
func parsePage(r *http.Request) int {
	pageValue := r.URL.Query().Get("page")
	page := 1
	if pageValue != "" {
		if parsed, err := strconv.Atoi(pageValue); err == nil && parsed > 0 {
			page = parsed
		}
	}
	return page
}

// formatTime formats a time value for display in templates.
// Returns empty string for zero times (unset timestamps).
// Format: YYYY-MM-DD HH:MM
func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}
