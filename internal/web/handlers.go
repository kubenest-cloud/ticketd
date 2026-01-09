package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ticketd/internal/config"
	"ticketd/internal/store"
)

const (
	pageSize = 20
)

type App struct {
	Store      store.Store
	Cfg        config.Config
	Templates  *templateCache
	DefaultCSS []byte
	AdminFS    fs.FS
}

func NewApp(cfg config.Config, st store.Store) (*App, error) {
	tmpl, err := parseTemplates()
	if err != nil {
		return nil, err
	}
	css, err := defaultCSS()
	if err != nil {
		return nil, err
	}
	adminFS, err := adminAssets()
	if err != nil {
		return nil, err
	}
	return &App{
		Store:      st,
		Cfg:        cfg,
		Templates:  tmpl,
		DefaultCSS: css,
		AdminFS:    adminFS,
	}, nil
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/admin/assets/*", http.StripPrefix("/admin/assets/", http.FileServer(http.FS(a.AdminFS))))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/form.css", a.handleFormCSS)
	r.Get("/embed/{formID}.js", a.handleEmbedJS)
	r.Options("/api/forms/{formID}/submit", a.handleSubmitOptions)
	r.Post("/api/forms/{formID}/submit", a.handleSubmit)

	r.Group(func(admin chi.Router) {
		admin.Use(a.basicAuth)
		admin.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/submissions", http.StatusFound)
		})
		admin.Get("/admin/submissions", a.handleAdminSubmissions)
		admin.Get("/admin/submissions/{submissionID}", a.handleAdminSubmissionView)
		admin.Post("/admin/submissions/{submissionID}/status", a.handleAdminUpdateSubmissionStatus)
		admin.Post("/admin/submissions/{submissionID}/delete", a.handleAdminDeleteSubmission)
		admin.Get("/admin/clients", a.handleAdminClients)
		admin.Post("/admin/clients", a.handleAdminCreateClient)
		admin.Get("/admin/clients/{clientID}/edit", a.handleAdminEditClient)
		admin.Post("/admin/clients/{clientID}/edit", a.handleAdminUpdateClient)
		admin.Get("/admin/clients/{clientID}/forms", a.handleAdminForms)
		admin.Post("/admin/clients/{clientID}/forms", a.handleAdminCreateForm)
	})

	return r
}

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

func (a *App) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if debugEnabled() {
		log.Printf("submit start form_id=%s origin=%q referer=%q content_type=%q", chi.URLParam(r, "formID"), r.Header.Get("Origin"), r.Header.Get("Referer"), r.Header.Get("Content-Type"))
	}
	allowed, origin := a.checkAllowedOrigin(r)
	if !allowed {
		if debugEnabled() {
			log.Printf("submit blocked form_id=%s origin=%q referer=%q", chi.URLParam(r, "formID"), r.Header.Get("Origin"), r.Header.Get("Referer"))
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden domain"})
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

func (a *App) handleAdminSubmissions(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	offset := (page - 1) * pageSize

	subs, total, err := a.Store.ListSubmissions(offset, pageSize)
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

	data := submissionsPage{
		Active:      "submissions",
		Submissions: items,
		Page:        page,
		Total:       total,
		TotalPages:  totalPages(total),
		PrevPage:    prevPage(page),
		NextPage:    nextPage(page, total),
	}

	a.renderTemplate(w, r, "submissions.html", data)
}

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

func (a *App) baseURLForAdmin(r *http.Request) (string, string) {
	if a.Cfg.PublicBaseURL != "" {
		return strings.TrimRight(a.Cfg.PublicBaseURL, "/"), ""
	}
	return a.publicBaseURL(r), "Set TICKETD_PUBLIC_BASE_URL in production for stable embed links."
}

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

func domainAllowed(host, allowed string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	allowed = strings.ToLower(strings.TrimSpace(allowed))
	if host == "" || allowed == "" {
		return false
	}
	if host == allowed {
		return true
	}
	return strings.HasSuffix(host, "."+allowed)
}

func buildEmbedJS(form store.Form, client store.Client, baseURL string) (string, error) {
	cssURL := fmt.Sprintf("%s/form.css", baseURL)
	apiURL := fmt.Sprintf("%s/api/forms/%d/submit", baseURL, form.ID)
	formTitle := fmt.Sprintf("%s - %s", client.Name, form.Name)

	fields := []map[string]any{
		{"label": "Name", "name": "name", "type": "text"},
		{"label": "Email", "name": "email", "type": "email"},
	}
	if form.Type == store.FormTypeSupport {
		fields = append(fields, map[string]any{"label": "Subject", "name": "subject", "type": "text"})
		fields = append(fields, map[string]any{
			"label":   "Priority",
			"name":    "priority",
			"type":    "select",
			"options": []string{"low", "medium", "high"},
		})
	}
	fields = append(fields, map[string]any{"label": "Message", "name": "message", "type": "textarea"})

	payload := map[string]any{
		"cssURL":   cssURL,
		"apiURL":   apiURL,
		"title":    formTitle,
		"fields":   fields,
		"formType": string(form.Type),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	script := fmt.Sprintf(`(function(){
  var cfg = %s;
  var scriptTag = document.currentScript;
  var mount = document.createElement("div");
  mount.className = "ticketd-embed";
  if (scriptTag && scriptTag.parentNode) {
    scriptTag.parentNode.insertBefore(mount, scriptTag);
  } else {
    document.body.appendChild(mount);
  }
  if (!document.querySelector('link[data-ticketd="true"]')) {
    var link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = cfg.cssURL;
    link.setAttribute("data-ticketd", "true");
    document.head.appendChild(link);
  }

  var form = document.createElement("form");
  form.className = "ticketd-form";
  var title = document.createElement("h3");
  title.textContent = cfg.title;
  form.appendChild(title);

  cfg.fields.forEach(function(field){
    var label = document.createElement("label");
    label.textContent = field.label;
    var input;
    if (field.type === "textarea") {
      input = document.createElement("textarea");
      input.rows = 4;
    } else if (field.type === "select") {
      input = document.createElement("select");
      field.options.forEach(function(opt){
        var option = document.createElement("option");
        option.value = opt;
        option.textContent = opt;
        input.appendChild(option);
      });
    } else {
      input = document.createElement("input");
      input.type = field.type || "text";
    }
    input.name = field.name;
    input.required = true;
    form.appendChild(label);
    form.appendChild(input);
  });

  var button = document.createElement("button");
  button.type = "submit";
  button.textContent = "Send";
  form.appendChild(button);

  var status = document.createElement("div");
  status.className = "ticketd-status";
  form.appendChild(status);

  form.addEventListener("submit", function(event){
    event.preventDefault();
    status.textContent = "Sending...";
    status.className = "ticketd-status";
    var payload = {};
    Array.prototype.forEach.call(form.elements, function(el){
      if (!el.name || el.type === "submit") {
        return;
      }
      payload[el.name] = el.value;
    });
    fetch(cfg.apiURL, {
      method: "POST",
      mode: "cors",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    })
      .then(function(res){ return res.json().then(function(body){ return { ok: res.ok, body: body }; }); })
      .then(function(result){
        if (!result.ok) {
          throw new Error(result.body && result.body.error ? result.body.error : "Failed");
        }
        status.textContent = "Thanks! We'll be in touch.";
        status.className = "ticketd-status ticketd-success";
        form.reset();
      })
      .catch(function(err){
        status.textContent = err.message || "Failed to send.";
        status.className = "ticketd-status ticketd-error";
      });
  });

  mount.appendChild(form);
})();`, string(data))

	return script, nil
}

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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

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

func debugEnabled() bool {
	return os.Getenv("TICKETD_DEBUG") == "1"
}

func parseID(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

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

func totalPages(total int) int {
	if total == 0 {
		return 1
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}

func prevPage(current int) int {
	if current > 1 {
		return current - 1
	}
	return 0
}

func nextPage(current, total int) int {
	if current < totalPages(total) {
		return current + 1
	}
	return 0
}

func isValidStatus(status string) bool {
	switch status {
	case "OPEN", "IN PROGRESS", "CLOSED":
		return true
	default:
		return false
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04")
}

type submissionView struct {
	store.Submission
	CreatedAt string
	FormType  string
}

type submissionsPage struct {
	Active      string
	Submissions []submissionView
	Page        int
	Total       int
	TotalPages  int
	PrevPage    int
	NextPage    int
}

type submissionPage struct {
	Active     string
	Submission store.Submission
	CreatedAt  string
}

type clientView struct {
	store.Client
	CreatedAt string
}

type clientsPage struct {
	Active     string
	Clients    []clientView
	Page       int
	Total      int
	TotalPages int
	PrevPage   int
	NextPage   int
}

type clientEditPage struct {
	Active string
	Client clientView
}

type formView struct {
	store.Form
	CreatedAt string
}

type formsPage struct {
	Active      string
	Client      clientView
	Forms       []formView
	BaseURL     string
	BaseURLNote string
}
