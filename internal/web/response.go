package web

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// renderTemplate renders a template page with the provided data.
// It executes the template with the "layout" base template and writes the result to the response.
// Returns a 500 error if the template is not found or fails to execute.
func (a *App) renderTemplate(w http.ResponseWriter, r *http.Request, page string, data any) {
	tmpl, ok := a.Templates.pages[page]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Printf("template error (%s): %v", page, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// writeJSON writes a JSON response with the given status code and payload.
// It sets the Content-Type header to application/json and encodes the payload.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
