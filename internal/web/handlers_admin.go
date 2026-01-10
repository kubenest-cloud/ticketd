package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ticketd/internal/store"
)

// handleAdminSubmissions displays a paginated, filterable list of form submissions.
// Supports filtering by status, client, form, and subject search.
// Submissions without a status are defaulted to "OPEN".
func (a *App) handleAdminSubmissions(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	offset := (page - 1) * pageSize

	// Parse filter parameters
	status := r.URL.Query().Get("status")
	clientID, _ := parseID(r.URL.Query().Get("client"))
	formID, _ := parseID(r.URL.Query().Get("form"))
	subjectSearch := strings.TrimSpace(r.URL.Query().Get("search"))

	// Use filtering if any filters are provided
	var subs []store.Submission
	var total int
	var err error

	hasFilters := status != "" || clientID > 0 || formID > 0 || subjectSearch != ""
	if hasFilters {
		subs, total, err = a.Store.FilterSubmissions(offset, pageSize, status, clientID, formID, subjectSearch)
	} else {
		subs, total, err = a.Store.ListSubmissions(offset, pageSize)
	}

	if err != nil {
		http.Error(w, "failed to load submissions", http.StatusInternalServerError)
		return
	}

	items := make([]submissionView, 0, len(subs))
	for _, sub := range subs {
		if sub.Status == "" {
			sub.Status = "OPEN"
		}
		items = append(items, submissionView{
			Submission: sub,
			CreatedAt:  formatTime(sub.CreatedAt),
			FormType:   string(sub.FormType),
		})
	}

	// Get clients and forms for filter dropdowns
	clients, _, _ := a.Store.ListClients(0, 1000) // Get all clients
	allForms := []store.Form{}
	for _, client := range clients {
		forms, _ := a.Store.ListForms(client.ID)
		allForms = append(allForms, forms...)
	}

	data := submissionsPage{
		Active:        "submissions",
		Submissions:   items,
		Page:          page,
		Total:         total,
		TotalPages:    totalPages(total),
		PrevPage:      prevPage(page),
		NextPage:      nextPage(page, total),
		Clients:       clients,
		Forms:         allForms,
		FilterStatus:  status,
		FilterClient:  clientID,
		FilterForm:    formID,
		FilterSearch:  subjectSearch,
		HasFilters:    hasFilters,
		ResultsCount:  len(subs),
	}

	a.renderTemplate(w, r, "submissions.html", data)
}

// handleAdminSubmissionView displays the details of a single submission.
// It shows all submission fields and allows updating the status or deleting the submission.
func (a *App) handleAdminSubmissionView(w http.ResponseWriter, r *http.Request) {
	submissionID, err := parseID(chi.URLParam(r, "submissionID"))
	if err != nil {
		http.Error(w, "invalid submission", http.StatusBadRequest)
		return
	}
	submission, err := a.Store.GetSubmission(submissionID)
	if err != nil {
		http.Error(w, "submission not found", http.StatusNotFound)
		return
	}
	if submission.Status == "" {
		submission.Status = "OPEN"
	}
	data := submissionPage{
		Active:     "submissions",
		Submission: submission,
		CreatedAt:  formatTime(submission.CreatedAt),
	}
	a.renderTemplate(w, r, "submission.html", data)
}

// handleAdminUpdateSubmissionStatus updates the status of a submission.
// Valid statuses are: OPEN, IN_PROGRESS, CLOSED (note: IN_PROGRESS not "IN PROGRESS").
// Redirects back to the submission view page after successful update.
func (a *App) handleAdminUpdateSubmissionStatus(w http.ResponseWriter, r *http.Request) {
	submissionID, err := parseID(chi.URLParam(r, "submissionID"))
	if err != nil {
		http.Error(w, "invalid submission", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	status := strings.ToUpper(strings.TrimSpace(r.FormValue("status")))
	if !isValidStatus(status) {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	if err := a.Store.UpdateSubmissionStatus(submissionID, status); err != nil {
		http.Error(w, "failed to update status", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/submissions/%d", submissionID), http.StatusFound)
}

// handleAdminDeleteSubmission deletes a submission permanently.
// Redirects back to the submissions list after successful deletion.
func (a *App) handleAdminDeleteSubmission(w http.ResponseWriter, r *http.Request) {
	submissionID, err := parseID(chi.URLParam(r, "submissionID"))
	if err != nil {
		http.Error(w, "invalid submission", http.StatusBadRequest)
		return
	}
	if err := a.Store.DeleteSubmission(submissionID); err != nil {
		http.Error(w, "failed to delete submission", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/submissions", http.StatusFound)
}

// isValidStatus checks if a status string is one of the valid submission statuses.
// Note: The validator package uses IN_PROGRESS (with underscore), not "IN PROGRESS".
func isValidStatus(status string) bool {
	switch status {
	case "OPEN", "IN_PROGRESS", "CLOSED":
		return true
	default:
		return false
	}
}

// submissionView is a view model for rendering submission list items.
// It includes formatted timestamps and form type for display.
type submissionView struct {
	store.Submission
	CreatedAt string
	FormType  string
}

// submissionsPage is the data structure for the submissions list page.
// It includes pagination information, filter options, and the list of submissions.
type submissionsPage struct {
	Active        string
	Submissions   []submissionView
	Page          int
	Total         int
	TotalPages    int
	PrevPage      int
	NextPage      int
	Clients       []store.Client
	Forms         []store.Form
	FilterStatus  string
	FilterClient  int64
	FilterForm    int64
	FilterSearch  string
	HasFilters    bool
	ResultsCount  int
}

// submissionPage is the data structure for the single submission detail page.
type submissionPage struct {
	Active     string
	Submission store.Submission
	CreatedAt  string
}
