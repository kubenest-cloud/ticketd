package web

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ticketd/internal/store"
)

// handleAdminClients displays a paginated list of all clients.
// Each client represents an organization that can create forms.
func (a *App) handleAdminClients(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	offset := (page - 1) * pageSize

	clients, total, err := a.Store.ListClients(offset, pageSize)
	if err != nil {
		http.Error(w, "failed to load clients", http.StatusInternalServerError)
		return
	}

	views := make([]clientView, 0, len(clients))
	for _, c := range clients {
		views = append(views, clientView{Client: c, CreatedAt: formatTime(c.CreatedAt)})
	}

	data := clientsPage{
		Active:     "clients",
		Clients:    views,
		Page:       page,
		Total:      total,
		TotalPages: totalPages(total),
		PrevPage:   prevPage(page),
		NextPage:   nextPage(page, total),
	}

	a.renderTemplate(w, r, "clients.html", data)
}

// handleAdminCreateClient creates a new client with the given name and allowed domain.
// The allowed domain is used for CORS validation when forms are submitted.
// Redirects back to the clients list after successful creation.
func (a *App) handleAdminCreateClient(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	domain := strings.TrimSpace(r.FormValue("allowed_domain"))
	if name == "" || domain == "" {
		http.Error(w, "name and allowed domain required", http.StatusBadRequest)
		return
	}
	if _, err := a.Store.CreateClient(name, domain); err != nil {
		http.Error(w, "failed to create client", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/clients", http.StatusFound)
}

// handleAdminEditClient displays the edit form for a specific client.
// Shows the current values for the client's name and allowed domain.
func (a *App) handleAdminEditClient(w http.ResponseWriter, r *http.Request) {
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
	data := clientEditPage{
		Active: "clients",
		Client: clientView{Client: client, CreatedAt: formatTime(client.CreatedAt)},
	}
	a.renderTemplate(w, r, "client_edit.html", data)
}

// handleAdminUpdateClient updates an existing client's name and allowed domain.
// Redirects back to the clients list after successful update.
func (a *App) handleAdminUpdateClient(w http.ResponseWriter, r *http.Request) {
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
	domain := strings.TrimSpace(r.FormValue("allowed_domain"))
	if name == "" || domain == "" {
		http.Error(w, "name and allowed domain required", http.StatusBadRequest)
		return
	}
	if err := a.Store.UpdateClient(clientID, name, domain); err != nil {
		http.Error(w, "failed to update client", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/clients", http.StatusFound)
}

// handleAdminDeleteClient deletes a client and all associated forms and submissions.
func (a *App) handleAdminDeleteClient(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseID(chi.URLParam(r, "clientID"))
	if err != nil {
		http.Error(w, "invalid client", http.StatusBadRequest)
		return
	}

	if err := a.Store.DeleteClient(clientID); err != nil {
		http.Error(w, "failed to delete client", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/clients", http.StatusFound)
}

// clientView is a view model for rendering client information.
// It includes a formatted timestamp for display in templates.
type clientView struct {
	store.Client
	CreatedAt string
}

// clientsPage is the data structure for the clients list page.
// It includes pagination information and the list of clients.
type clientsPage struct {
	Active     string
	Clients    []clientView
	Page       int
	Total      int
	TotalPages int
	PrevPage   int
	NextPage   int
}

// clientEditPage is the data structure for the client edit page.
type clientEditPage struct {
	Active string
	Client clientView
}
