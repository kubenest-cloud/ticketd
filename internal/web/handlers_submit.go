package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"ticketd/internal/store"
)

// handleSubmitOptions handles CORS preflight requests for form submissions.
// It checks if the origin is allowed based on the client's allowed domain.
// Returns 403 Forbidden if the origin is not allowed, or 204 No Content with CORS headers if allowed.
func (a *App) handleSubmitOptions(w http.ResponseWriter, r *http.Request) {
	if debugEnabled() {
		log.Printf("preflight form_id=%s origin=%q referer=%q", chi.URLParam(r, "formID"), r.Header.Get("Origin"), r.Header.Get("Referer"))
	}
	allowed, origin := a.checkAllowedOrigin(r)
	if !allowed {
		if debugEnabled() {
			log.Printf("preflight blocked form_id=%s origin=%q referer=%q", chi.URLParam(r, "formID"), r.Header.Get("Origin"), r.Header.Get("Referer"))
		}
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusNoContent)
}

// handleSubmit processes form submissions from embedded forms.
// It validates the origin, parses the submission data (JSON or form-encoded),
// validates the input, stores the submission, and returns a JSON response.
// Supports both application/json and application/x-www-form-urlencoded content types.
func (a *App) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if debugEnabled() {
		log.Printf("submit start form_id=%s origin=%q referer=%q content_type=%q", chi.URLParam(r, "formID"), r.Header.Get("Origin"), r.Header.Get("Referer"), r.Header.Get("Content-Type"))
	}
	allowed, origin := a.checkAllowedOrigin(r)
	if !allowed {
		// Get more details for better error message
		formID, _ := parseID(chi.URLParam(r, "formID"))
		form, err := a.Store.GetForm(formID)
		var allowedDomain string
		if err == nil {
			if client, err := a.Store.GetClient(form.ClientID); err == nil {
				allowedDomain = client.AllowedDomain
			}
		}

		if debugEnabled() {
			log.Printf("submit blocked form_id=%s origin=%q referer=%q allowed_domain=%q", chi.URLParam(r, "formID"), r.Header.Get("Origin"), r.Header.Get("Referer"), allowedDomain)
		}

		// Provide helpful error message in development
		errorMsg := "forbidden domain"
		if allowedDomain != "" {
			errorMsg = fmt.Sprintf("domain not allowed - configure client allowed domain to match your site (currently set to: %s)", allowedDomain)
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": errorMsg})
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}

	formID, err := parseID(chi.URLParam(r, "formID"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
		return
	}
	form, err := a.Store.GetForm(formID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "form not found"})
		return
	}

	input := store.SubmissionInput{
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var payload struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Subject  string `json:"subject"`
			Message  string `json:"message"`
			Priority string `json:"priority"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		input.Name = strings.TrimSpace(payload.Name)
		input.Email = strings.TrimSpace(payload.Email)
		input.Subject = strings.TrimSpace(payload.Subject)
		input.Message = strings.TrimSpace(payload.Message)
		input.Priority = strings.TrimSpace(payload.Priority)
		if debugEnabled() {
			log.Printf("submit json form_id=%d name=%q email=%q subject=%q priority=%q message_len=%d", form.ID, input.Name, input.Email, input.Subject, input.Priority, len(input.Message))
		}
	} else {
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
			return
		}
		input.Name = strings.TrimSpace(formValue(r, "name"))
		input.Email = strings.TrimSpace(formValue(r, "email"))
		input.Subject = strings.TrimSpace(formValue(r, "subject"))
		input.Message = strings.TrimSpace(formValue(r, "message"))
		input.Priority = strings.TrimSpace(formValue(r, "priority"))
		if debugEnabled() {
			log.Printf("submit form form_id=%d name=%q email=%q subject=%q priority=%q message_len=%d content_type=%q", form.ID, input.Name, input.Email, input.Subject, input.Priority, len(input.Message), contentType)
		}
	}

	if err := validateSubmission(form.Type, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if _, err := a.Store.CreateSubmission(form.ID, input); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// checkAllowedOrigin validates if the request origin is allowed to submit to this form.
// It checks the Origin header first, then falls back to the Referer header.
// Returns true and the origin if allowed, or false and empty string if not allowed.
// The origin is matched against the client's allowed domain (exact match or subdomain).
func (a *App) checkAllowedOrigin(r *http.Request) (bool, string) {
	origin := r.Header.Get("Origin")
	referer := r.Header.Get("Referer")
	var host string
	if origin != "" {
		if parsed, err := url.Parse(origin); err == nil {
			host = parsed.Hostname()
		}
	} else if referer != "" {
		if parsed, err := url.Parse(referer); err == nil {
			host = parsed.Hostname()
		}
	}
	if host == "" {
		return false, ""
	}

	formID, err := parseID(chi.URLParam(r, "formID"))
	if err != nil {
		return false, ""
	}
	form, err := a.Store.GetForm(formID)
	if err != nil {
		return false, ""
	}
	client, err := a.Store.GetClient(form.ClientID)
	if err != nil {
		return false, ""
	}
	if !domainAllowed(host, client.AllowedDomain) {
		return false, ""
	}
	return true, origin
}

// domainAllowed checks if a host matches or is a subdomain of the allowed domain.
// For example, if allowed is "example.com", it will match "example.com" and "www.example.com".
// Special handling for localhost: "localhost" will match "localhost:3000", "localhost:8080", etc.
func domainAllowed(host, allowed string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	allowed = strings.ToLower(strings.TrimSpace(allowed))
	if host == "" || allowed == "" {
		return false
	}

	// Strip port from localhost and 127.0.0.1 for easier development
	// This allows "localhost" to match "localhost:3000", "localhost:5173", etc.
	// Also allows "127.0.0.1" to match "127.0.0.1:3000", etc.
	if strings.HasPrefix(host, "localhost:") {
		host = "localhost"
	}
	if strings.HasPrefix(allowed, "localhost:") {
		allowed = "localhost"
	}
	if strings.HasPrefix(host, "127.0.0.1:") {
		host = "127.0.0.1"
	}
	if strings.HasPrefix(allowed, "127.0.0.1:") {
		allowed = "127.0.0.1"
	}
	// Allow localhost and 127.0.0.1 to be interchangeable
	if (host == "localhost" && allowed == "127.0.0.1") || (host == "127.0.0.1" && allowed == "localhost") {
		return true
	}

	if host == allowed {
		return true
	}
	return strings.HasSuffix(host, "."+allowed)
}

// validateSubmission validates form submission input based on the form type.
// Support forms require subject and priority, contact forms require name and email.
// Basic email format validation is performed if email is provided.
func validateSubmission(formType store.FormType, input *store.SubmissionInput) error {
	if input.Message == "" {
		return fmt.Errorf("message is required")
	}
	switch formType {
	case store.FormTypeSupport:
		if input.Subject == "" {
			return fmt.Errorf("subject is required")
		}
		if input.Priority == "" {
			input.Priority = "medium"
		}
	case store.FormTypeContact:
		if input.Name == "" || input.Email == "" {
			return fmt.Errorf("name and email are required")
		}
	default:
		return fmt.Errorf("invalid form type")
	}
	if input.Email != "" && !strings.Contains(input.Email, "@") {
		return fmt.Errorf("invalid email")
	}
	return nil
}

// formValue retrieves a form value from either regular form data or multipart form data.
// This handles both application/x-www-form-urlencoded and multipart/form-data submissions.
func formValue(r *http.Request, key string) string {
	if value := r.FormValue(key); value != "" {
		return value
	}
	if r.MultipartForm != nil {
		if values, ok := r.MultipartForm.Value[key]; ok && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
