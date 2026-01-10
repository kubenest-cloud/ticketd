package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ticketd/internal/store"
)

// handleAdminForms displays all forms for a specific client.
// Each form has an embed code that can be copied and pasted into websites.
// The base URL for embed codes is taken from the config or inferred from the request.
func (a *App) handleAdminForms(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseID(chi.URLParam(r, "clientID"))
	if err != nil {
		http.Error(w, "invalid client", http.StatusBadRequest)
		return
	}
	client, err := a.Store.GetClient(clientID)
	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	forms, err := a.Store.ListForms(clientID)
	if err != nil {
		http.Error(w, "failed to load forms", http.StatusInternalServerError)
		return
	}

	views := make([]formView, 0, len(forms))
	for _, f := range forms {
		views = append(views, formView{Form: f, CreatedAt: formatTime(f.CreatedAt)})
	}

	baseURL, note := a.baseURLForAdmin(r)
	data := formsPage{
		Active:      "clients",
		Client:      clientView{Client: client, CreatedAt: formatTime(client.CreatedAt)},
		Forms:       views,
		BaseURL:     baseURL,
		BaseURLNote: note,
	}
	a.renderTemplate(w, r, "forms.html", data)
}

// handleAdminCreateForm creates a new form for a client.
// Forms can be of type "contact" or "support", which determines the required fields.
// Redirects back to the forms list after successful creation.
func (a *App) handleAdminCreateForm(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseID(chi.URLParam(r, "clientID"))
	if err != nil {
		http.Error(w, "invalid client", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	typeValue := strings.TrimSpace(r.FormValue("type"))
	formType := store.FormType(typeValue)
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if _, err := a.Store.CreateForm(clientID, name, formType); err != nil {
		http.Error(w, "failed to create form", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/clients/%d/forms", clientID), http.StatusFound)
}

// handleAdminEditFormPage displays the form edit page.
func (a *App) handleAdminEditFormPage(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseID(chi.URLParam(r, "clientID"))
	if err != nil {
		http.Error(w, "invalid client", http.StatusBadRequest)
		return
	}
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

	// Verify form belongs to the client
	if form.ClientID != clientID {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	data := formEditPage{
		Active:   "clients",
		ClientID: clientID,
		Form:     form,
	}
	a.renderTemplate(w, r, "form_edit.html", data)
}

// handleAdminUpdateForm updates an existing form.
func (a *App) handleAdminUpdateForm(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseID(chi.URLParam(r, "clientID"))
	if err != nil {
		http.Error(w, "invalid client", http.StatusBadRequest)
		return
	}
	formID, err := parseID(chi.URLParam(r, "formID"))
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	typeValue := strings.TrimSpace(r.FormValue("type"))
	formType := store.FormType(typeValue)

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	// Verify form belongs to the client
	form, err := a.Store.GetForm(formID)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}
	if form.ClientID != clientID {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	if err := a.Store.UpdateForm(formID, name, formType); err != nil {
		http.Error(w, "failed to update form", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/clients/%d/forms", clientID), http.StatusFound)
}

// handleAdminDeleteForm deletes a form and all associated submissions.
func (a *App) handleAdminDeleteForm(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseID(chi.URLParam(r, "clientID"))
	if err != nil {
		http.Error(w, "invalid client", http.StatusBadRequest)
		return
	}
	formID, err := parseID(chi.URLParam(r, "formID"))
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	// Verify form belongs to the client
	form, err := a.Store.GetForm(formID)
	if err != nil {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}
	if form.ClientID != clientID {
		http.Error(w, "form not found", http.StatusNotFound)
		return
	}

	if err := a.Store.DeleteForm(formID); err != nil {
		http.Error(w, "failed to delete form", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/clients/%d/forms", clientID), http.StatusFound)
}

// formView is a view model for rendering form information.
// It includes a formatted timestamp for display in templates.
type formView struct {
	store.Form
	CreatedAt string
}

// formsPage is the data structure for the forms list page.
// It includes the parent client, the list of forms, and base URL information for embed codes.
type formsPage struct {
	Active      string
	Client      clientView
	Forms       []formView
	BaseURL     string
	BaseURLNote string
}

// formEditPage is the data structure for the form edit page.
type formEditPage struct {
	Active   string
	ClientID int64
	Form     store.Form
}
