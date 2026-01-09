package web

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

// handleFormCSS serves the CSS stylesheet for embedded forms.
// If a custom CSS path is configured and the file exists, it serves that.
// Otherwise, it serves the default embedded CSS.
func (a *App) handleFormCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if a.Cfg.CustomCSSPath != "" {
		data, err := os.ReadFile(a.Cfg.CustomCSSPath)
		if err == nil {
			_, _ = w.Write(data)
			return
		}
	}
	_, _ = w.Write(a.DefaultCSS)
}

// handleEmbedJS generates and serves the JavaScript embed code for a specific form.
// The JavaScript creates a self-contained form widget that can be embedded on any website.
// It handles CORS validation based on the client's allowed domain.
func (a *App) handleEmbedJS(w http.ResponseWriter, r *http.Request) {
	formID, err := parseID(chi.URLParam(r, "formID"))
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	form, err := a.Store.GetForm(formID)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}
	client, err := a.Store.GetClient(form.ClientID)
	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}

	baseURL := a.publicBaseURL(r)
	js, err := buildEmbedJS(form, client, baseURL)
	if err != nil {
		http.Error(w, "script error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write([]byte(js))
}
