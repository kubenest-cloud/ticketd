package web

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ticketd/internal/config"
	"ticketd/internal/store"
)

// App holds the application dependencies and state.
// It is the main entry point for the web layer and contains
// the store, configuration, templates, and static assets.
type App struct {
	Store      store.Store
	Cfg        config.Config
	Templates  *templateCache
	DefaultCSS []byte
	AdminFS    fs.FS
}

// NewApp creates a new App instance with all dependencies initialized.
// It loads templates, default CSS, and admin assets.
// Returns an error if any initialization fails.
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

// Router creates and configures the HTTP router with all application routes.
// It sets up middleware, public endpoints, and protected admin routes.
func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Static assets for admin interface
	r.Handle("/admin/assets/*", http.StripPrefix("/admin/assets/", http.FileServer(http.FS(a.AdminFS))))

	// Public endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/embed/form.css", a.handleFormCSS)
	r.Get("/embed/{formID}.js", a.handleEmbedJS)
	r.Options("/api/forms/{formID}/submit", a.handleSubmitOptions)
	r.Post("/api/forms/{formID}/submit", a.handleSubmit)

	// Protected admin routes
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
		admin.Post("/admin/clients/{clientID}/delete", a.handleAdminDeleteClient)
		admin.Get("/admin/clients/{clientID}/forms", a.handleAdminForms)
		admin.Post("/admin/clients/{clientID}/forms", a.handleAdminCreateForm)
		admin.Get("/admin/clients/{clientID}/forms/{formID}/edit", a.handleAdminEditFormPage)
		admin.Post("/admin/clients/{clientID}/forms/{formID}/edit", a.handleAdminUpdateForm)
		admin.Post("/admin/clients/{clientID}/forms/{formID}/delete", a.handleAdminDeleteForm)
	})

	return r
}
