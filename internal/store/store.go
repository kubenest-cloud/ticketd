// Package store defines the data models and persistence interface for TicketD.
// It uses a repository pattern to allow swapping database implementations
// while maintaining a consistent API for data access.
package store

import "time"

// Client represents a client organization that can create forms.
// Each client has an allowed domain used for CORS validation of form submissions.
type Client struct {
	ID            int64
	Name          string
	AllowedDomain string
	CreatedAt     time.Time
}

// FormType represents the type of form (support or contact).
type FormType string

const (
	// FormTypeSupport represents a support form with name, email, subject, message, and priority fields.
	FormTypeSupport FormType = "support"

	// FormTypeContact represents a contact form with name, email, subject, and message fields.
	FormTypeContact FormType = "contact"
)

// Form represents a contact or support form belonging to a client.
type Form struct {
	ID        int64
	ClientID  int64
	Name      string
	Type      FormType
	CreatedAt time.Time
}

// Submission represents a form submission (ticket).
// It includes denormalized client and form names for easier display.
type Submission struct {
	ID        int64
	ClientID  int64
	Client    string   // Denormalized client name
	FormID    int64
	Form      string   // Denormalized form name
	FormType  FormType
	Status    string
	Name      string
	Email     string
	Subject   string
	Message   string
	Priority  string
	IP        string
	UserAgent string
	CreatedAt time.Time
}

// SubmissionInput contains the data needed to create a new submission.
type SubmissionInput struct {
	Name      string
	Email     string
	Subject   string
	Message   string
	Priority  string
	IP        string
	UserAgent string
}

// Store defines the persistence interface for all data operations.
// Implementations must provide ACID guarantees for data integrity.
type Store interface {
	// Migrate runs database migrations to ensure schema is up to date.
	Migrate() error

	// Close closes the database connection and releases resources.
	Close() error

	// CreateClient creates a new client with the given name and allowed domain.
	// The allowed domain is used for CORS validation of form submissions.
	// Returns the created client or an error if creation fails.
	CreateClient(name, allowedDomain string) (Client, error)

	// ListClients returns a paginated list of clients and the total count.
	// offset specifies how many records to skip, limit specifies max records to return.
	ListClients(offset, limit int) ([]Client, int, error)

	// GetClient retrieves a client by ID.
	// Returns ErrNotFound if the client doesn't exist.
	GetClient(id int64) (Client, error)

	// UpdateClient updates an existing client's name and allowed domain.
	// Returns an error if the client doesn't exist or update fails.
	UpdateClient(id int64, name, allowedDomain string) error

	// DeleteClient permanently deletes a client and all associated forms and submissions.
	// Returns an error if the client doesn't exist or deletion fails.
	DeleteClient(id int64) error

	// CreateForm creates a new form for the specified client.
	// Returns the created form or an error if creation fails.
	CreateForm(clientID int64, name string, formType FormType) (Form, error)

	// ListForms returns all forms for the specified client.
	ListForms(clientID int64) ([]Form, error)

	// GetForm retrieves a form by ID.
	// Returns ErrNotFound if the form doesn't exist.
	GetForm(id int64) (Form, error)

	// UpdateForm updates an existing form's name and type.
	// Returns an error if the form doesn't exist or update fails.
	UpdateForm(id int64, name string, formType FormType) error

	// DeleteForm permanently deletes a form and all associated submissions.
	// Returns an error if the form doesn't exist or deletion fails.
	DeleteForm(id int64) error

	// CreateSubmission creates a new submission for the specified form.
	// Returns the created submission with denormalized client and form data.
	CreateSubmission(formID int64, input SubmissionInput) (Submission, error)

	// ListSubmissions returns a paginated list of submissions and the total count.
	// Results include denormalized client and form names for display.
	// offset specifies how many records to skip, limit specifies max records to return.
	ListSubmissions(offset, limit int) ([]Submission, int, error)

	// FilterSubmissions returns a filtered paginated list of submissions and the total count.
	// Filters can be applied by status, client ID, form ID, and subject search.
	// Empty/zero values for filters are ignored (no filtering applied for that field).
	FilterSubmissions(offset, limit int, status string, clientID, formID int64, subjectSearch string) ([]Submission, int, error)

	// GetSubmission retrieves a submission by ID with denormalized client and form data.
	// Returns ErrNotFound if the submission doesn't exist.
	GetSubmission(id int64) (Submission, error)

	// UpdateSubmissionStatus updates the status of a submission.
	// Valid statuses are OPEN, IN_PROGRESS, and CLOSED.
	UpdateSubmissionStatus(id int64, status string) error

	// DeleteSubmission permanently deletes a submission.
	// Returns an error if the submission doesn't exist or deletion fails.
	DeleteSubmission(id int64) error
}
