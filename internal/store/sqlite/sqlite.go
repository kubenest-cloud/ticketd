// Package sqlite implements the Store interface using SQLite as the database.
// It provides persistent storage for clients, forms, and submissions.
package sqlite

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	apperrors "ticketd/internal/errors"
	"ticketd/internal/store"
	"ticketd/internal/validator"
)

// Store implements the store.Store interface using SQLite.
type Store struct {
	db *sql.DB
}

// New creates a new SQLite store at the specified path.
// It opens the database connection and verifies connectivity.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to open database")
	}
	if err := db.Ping(); err != nil {
		return nil, apperrors.Wrap(err, "failed to connect to database")
	}
	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return apperrors.Wrap(err, "failed to close database")
	}
	return nil
}

// Migrate runs database migrations to create or update the schema.
// It creates the necessary tables if they don't exist.
func (s *Store) Migrate() error {
	// Create tables with IF NOT EXISTS to make migrations idempotent
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS clients (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	allowed_domain TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forms (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	client_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(client_id) REFERENCES clients(id)
);

CREATE TABLE IF NOT EXISTS submissions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	client_id INTEGER NOT NULL,
	form_id INTEGER NOT NULL,
	status TEXT NOT NULL DEFAULT 'OPEN',
	name TEXT,
	email TEXT,
	subject TEXT,
	message TEXT,
	priority TEXT,
	ip TEXT,
	user_agent TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(client_id) REFERENCES clients(id),
	FOREIGN KEY(form_id) REFERENCES forms(id)
);
`)
	if err != nil {
		return apperrors.Wrap(err, "failed to run database migrations")
	}

	// Note: The status column was added in a migration.
	// Since we're using CREATE TABLE IF NOT EXISTS, existing tables
	// already have the status column. This ALTER TABLE is kept for
	// backwards compatibility but will fail silently on existing tables.
	_, err = s.db.Exec(`ALTER TABLE submissions ADD COLUMN status TEXT NOT NULL DEFAULT 'OPEN'`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return apperrors.Wrap(err, "failed to add status column")
	}

	return nil
}

// CreateClient creates a new client after validating the input.
func (s *Store) CreateClient(name, allowedDomain string) (store.Client, error) {
	// Validate and trim input
	name, allowedDomain, err := validator.TrimAndValidateClient(name, allowedDomain)
	if err != nil {
		return store.Client{}, err
	}

	result, err := s.db.Exec(`INSERT INTO clients (name, allowed_domain) VALUES (?, ?)`, name, allowedDomain)
	if err != nil {
		return store.Client{}, apperrors.Wrap(err, "failed to create client")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return store.Client{}, apperrors.Wrap(err, "failed to get client ID")
	}

	return s.GetClient(id)
}

// ListClients returns a paginated list of clients ordered by creation date (newest first).
func (s *Store) ListClients(offset, limit int) ([]store.Client, int, error) {
	// Apply default pagination limits
	limit = formatLimit(limit)
	offset = formatOffset(offset)

	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM clients`).Scan(&total); err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to count clients")
	}

	rows, err := s.db.Query(`SELECT id, name, allowed_domain, created_at FROM clients ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list clients")
	}
	defer rows.Close()

	clients := []store.Client{}
	for rows.Next() {
		var client store.Client
		var created string
		if err := rows.Scan(&client.ID, &client.Name, &client.AllowedDomain, &created); err != nil {
			return nil, 0, apperrors.Wrap(err, "failed to scan client row")
		}
		client.CreatedAt = parseTime(created)
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.Wrap(err, "error iterating client rows")
	}

	return clients, total, nil
}

// GetClient retrieves a client by ID.
func (s *Store) GetClient(id int64) (store.Client, error) {
	var client store.Client
	var created string
	row := s.db.QueryRow(`SELECT id, name, allowed_domain, created_at FROM clients WHERE id = ?`, id)
	if err := row.Scan(&client.ID, &client.Name, &client.AllowedDomain, &created); err != nil {
		if err == sql.ErrNoRows {
			return store.Client{}, apperrors.NotFoundError("client", id)
		}
		return store.Client{}, apperrors.Wrapf(err, "failed to get client %d", id)
	}
	client.CreatedAt = parseTime(created)
	return client, nil
}

// UpdateClient updates an existing client's name and allowed domain.
func (s *Store) UpdateClient(id int64, name, allowedDomain string) error {
	// Validate and trim input
	name, allowedDomain, err := validator.TrimAndValidateClient(name, allowedDomain)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(`UPDATE clients SET name = ?, allowed_domain = ? WHERE id = ?`, name, allowedDomain, id)
	if err != nil {
		return apperrors.Wrapf(err, "failed to update client %d", id)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to check rows affected")
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("client", id)
	}

	return nil
}

// CreateForm creates a new form after validating the input.
func (s *Store) CreateForm(clientID int64, name string, formType store.FormType) (store.Form, error) {
	// Validate input
	name = strings.TrimSpace(name)
	if err := validator.ValidateForm(name, formType); err != nil {
		return store.Form{}, err
	}

	// Verify client exists
	if _, err := s.GetClient(clientID); err != nil {
		return store.Form{}, apperrors.Wrapf(err, "client %d not found", clientID)
	}

	result, err := s.db.Exec(`INSERT INTO forms (client_id, name, type) VALUES (?, ?, ?)`, clientID, name, string(formType))
	if err != nil {
		return store.Form{}, apperrors.Wrap(err, "failed to create form")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return store.Form{}, apperrors.Wrap(err, "failed to get form ID")
	}

	return s.GetForm(id)
}

// ListForms returns all forms for a client ordered by creation date (newest first).
func (s *Store) ListForms(clientID int64) ([]store.Form, error) {
	rows, err := s.db.Query(`SELECT id, client_id, name, type, created_at FROM forms WHERE client_id = ? ORDER BY created_at DESC`, clientID)
	if err != nil {
		return nil, apperrors.Wrapf(err, "failed to list forms for client %d", clientID)
	}
	defer rows.Close()

	forms := []store.Form{}
	for rows.Next() {
		var form store.Form
		var created string
		if err := rows.Scan(&form.ID, &form.ClientID, &form.Name, &form.Type, &created); err != nil {
			return nil, apperrors.Wrap(err, "failed to scan form row")
		}
		form.CreatedAt = parseTime(created)
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, "error iterating form rows")
	}

	return forms, nil
}

// GetForm retrieves a form by ID.
func (s *Store) GetForm(id int64) (store.Form, error) {
	var form store.Form
	var created string
	row := s.db.QueryRow(`SELECT id, client_id, name, type, created_at FROM forms WHERE id = ?`, id)
	if err := row.Scan(&form.ID, &form.ClientID, &form.Name, &form.Type, &created); err != nil {
		if err == sql.ErrNoRows {
			return store.Form{}, apperrors.NotFoundError("form", id)
		}
		return store.Form{}, apperrors.Wrapf(err, "failed to get form %d", id)
	}
	form.CreatedAt = parseTime(created)
	return form, nil
}

// CreateSubmission creates a new submission after validating the input.
func (s *Store) CreateSubmission(formID int64, input store.SubmissionInput) (store.Submission, error) {
	// Trim and validate input
	input = validator.TrimSubmissionInput(input)
	if err := validator.ValidateSubmission(input); err != nil {
		return store.Submission{}, err
	}

	// Verify form exists and get client ID
	form, err := s.GetForm(formID)
	if err != nil {
		return store.Submission{}, apperrors.Wrapf(err, "form %d not found", formID)
	}

	result, err := s.db.Exec(`
INSERT INTO submissions (client_id, form_id, status, name, email, subject, message, priority, ip, user_agent)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, form.ClientID, form.ID, validator.StatusOpen, input.Name, input.Email, input.Subject, input.Message, input.Priority, input.IP, input.UserAgent)
	if err != nil {
		return store.Submission{}, apperrors.Wrap(err, "failed to create submission")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return store.Submission{}, apperrors.Wrap(err, "failed to get submission ID")
	}

	return s.GetSubmission(id)
}

// ListSubmissions returns a paginated list of submissions with denormalized client and form data.
func (s *Store) ListSubmissions(offset, limit int) ([]store.Submission, int, error) {
	// Apply default pagination limits
	limit = formatLimit(limit)
	offset = formatOffset(offset)

	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM submissions`).Scan(&total); err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to count submissions")
	}

	rows, err := s.db.Query(`
SELECT s.id, s.client_id, c.name, s.form_id, f.name, f.type, s.status, s.name, s.email, s.subject, s.message, s.priority, s.ip, s.user_agent, s.created_at
FROM submissions s
JOIN clients c ON c.id = s.client_id
JOIN forms f ON f.id = s.form_id
ORDER BY s.created_at DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list submissions")
	}
	defer rows.Close()

	submissions := []store.Submission{}
	for rows.Next() {
		var submission store.Submission
		var created string
		if err := rows.Scan(&submission.ID, &submission.ClientID, &submission.Client, &submission.FormID, &submission.Form, &submission.FormType, &submission.Status, &submission.Name, &submission.Email, &submission.Subject, &submission.Message, &submission.Priority, &submission.IP, &submission.UserAgent, &created); err != nil {
			return nil, 0, apperrors.Wrap(err, "failed to scan submission row")
		}
		submission.CreatedAt = parseTime(created)
		submissions = append(submissions, submission)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.Wrap(err, "error iterating submission rows")
	}

	return submissions, total, nil
}

// GetSubmission retrieves a submission by ID with denormalized client and form data.
func (s *Store) GetSubmission(id int64) (store.Submission, error) {
	row := s.db.QueryRow(`
SELECT s.id, s.client_id, c.name, s.form_id, f.name, f.type, s.status, s.name, s.email, s.subject, s.message, s.priority, s.ip, s.user_agent, s.created_at
FROM submissions s
JOIN clients c ON c.id = s.client_id
JOIN forms f ON f.id = s.form_id
WHERE s.id = ?
`, id)

	var submission store.Submission
	var created string
	if err := row.Scan(&submission.ID, &submission.ClientID, &submission.Client, &submission.FormID, &submission.Form, &submission.FormType, &submission.Status, &submission.Name, &submission.Email, &submission.Subject, &submission.Message, &submission.Priority, &submission.IP, &submission.UserAgent, &created); err != nil {
		if err == sql.ErrNoRows {
			return store.Submission{}, apperrors.NotFoundError("submission", id)
		}
		return store.Submission{}, apperrors.Wrapf(err, "failed to get submission %d", id)
	}
	submission.CreatedAt = parseTime(created)
	return submission, nil
}

// UpdateSubmissionStatus updates the status of a submission after validating it.
func (s *Store) UpdateSubmissionStatus(id int64, status string) error {
	// Validate status
	status = strings.TrimSpace(status)
	if err := validator.ValidateStatus(status); err != nil {
		return err
	}

	result, err := s.db.Exec(`UPDATE submissions SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return apperrors.Wrapf(err, "failed to update submission %d status", id)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to check rows affected")
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("submission", id)
	}

	return nil
}

// DeleteSubmission permanently deletes a submission.
func (s *Store) DeleteSubmission(id int64) error {
	result, err := s.db.Exec(`DELETE FROM submissions WHERE id = ?`, id)
	if err != nil {
		return apperrors.Wrapf(err, "failed to delete submission %d", id)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to check rows affected")
	}
	if rowsAffected == 0 {
		return apperrors.NotFoundError("submission", id)
	}

	return nil
}

// parseTime attempts to parse a timestamp string from SQLite.
// It tries multiple formats: SQLite datetime format and RFC3339.
// Returns zero time if parsing fails.
func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	// Try SQLite datetime format first (most common)
	parsed, err := time.Parse("2006-01-02 15:04:05", value)
	if err == nil {
		return parsed
	}

	// Try RFC3339 format as fallback
	parsed, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed
	}

	// Return zero time if all parsing attempts fail
	return time.Time{}
}

// formatLimit ensures limit is within valid bounds for pagination.
// Returns default page size (20) if limit is <= 0.
func formatLimit(limit int) int {
	const defaultPageSize = 20
	if limit <= 0 {
		return defaultPageSize
	}
	return limit
}

// formatOffset ensures offset is non-negative for pagination.
// Returns 0 if offset is negative.
func formatOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
