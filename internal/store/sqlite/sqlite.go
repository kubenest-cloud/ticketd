package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"ticketd/internal/store"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate() error {
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
		return err
	}
	_, err = s.db.Exec(`ALTER TABLE submissions ADD COLUMN status TEXT NOT NULL DEFAULT 'OPEN'`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

func (s *Store) CreateClient(name, allowedDomain string) (store.Client, error) {
	result, err := s.db.Exec(`INSERT INTO clients (name, allowed_domain) VALUES (?, ?)`, name, allowedDomain)
	if err != nil {
		return store.Client{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return store.Client{}, err
	}
	return s.GetClient(id)
}

func (s *Store) ListClients(offset, limit int) ([]store.Client, int, error) {
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM clients`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(`SELECT id, name, allowed_domain, created_at FROM clients ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	clients := []store.Client{}
	for rows.Next() {
		var c store.Client
		var created string
		if err := rows.Scan(&c.ID, &c.Name, &c.AllowedDomain, &created); err != nil {
			return nil, 0, err
		}
		c.CreatedAt = parseTime(created)
		clients = append(clients, c)
	}
	return clients, total, rows.Err()
}

func (s *Store) GetClient(id int64) (store.Client, error) {
	var c store.Client
	var created string
	row := s.db.QueryRow(`SELECT id, name, allowed_domain, created_at FROM clients WHERE id = ?`, id)
	if err := row.Scan(&c.ID, &c.Name, &c.AllowedDomain, &created); err != nil {
		return store.Client{}, err
	}
	c.CreatedAt = parseTime(created)
	return c, nil
}

func (s *Store) UpdateClient(id int64, name, allowedDomain string) error {
	_, err := s.db.Exec(`UPDATE clients SET name = ?, allowed_domain = ? WHERE id = ?`, name, allowedDomain, id)
	return err
}

func (s *Store) CreateForm(clientID int64, name string, formType store.FormType) (store.Form, error) {
	result, err := s.db.Exec(`INSERT INTO forms (client_id, name, type) VALUES (?, ?, ?)`, clientID, name, string(formType))
	if err != nil {
		return store.Form{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return store.Form{}, err
	}
	return s.GetForm(id)
}

func (s *Store) ListForms(clientID int64) ([]store.Form, error) {
	rows, err := s.db.Query(`SELECT id, client_id, name, type, created_at FROM forms WHERE client_id = ? ORDER BY created_at DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	forms := []store.Form{}
	for rows.Next() {
		var f store.Form
		var created string
		if err := rows.Scan(&f.ID, &f.ClientID, &f.Name, &f.Type, &created); err != nil {
			return nil, err
		}
		f.CreatedAt = parseTime(created)
		forms = append(forms, f)
	}
	return forms, rows.Err()
}

func (s *Store) GetForm(id int64) (store.Form, error) {
	var f store.Form
	var created string
	row := s.db.QueryRow(`SELECT id, client_id, name, type, created_at FROM forms WHERE id = ?`, id)
	if err := row.Scan(&f.ID, &f.ClientID, &f.Name, &f.Type, &created); err != nil {
		return store.Form{}, err
	}
	f.CreatedAt = parseTime(created)
	return f, nil
}

func (s *Store) CreateSubmission(formID int64, input store.SubmissionInput) (store.Submission, error) {
	form, err := s.GetForm(formID)
	if err != nil {
		return store.Submission{}, err
	}
	result, err := s.db.Exec(`
INSERT INTO submissions (client_id, form_id, status, name, email, subject, message, priority, ip, user_agent)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, form.ClientID, form.ID, "OPEN", input.Name, input.Email, input.Subject, input.Message, input.Priority, input.IP, input.UserAgent)
	if err != nil {
		return store.Submission{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return store.Submission{}, err
	}
	return s.GetSubmission(id)
}

func (s *Store) ListSubmissions(offset, limit int) ([]store.Submission, int, error) {
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM submissions`).Scan(&total); err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}
	defer rows.Close()

	subs := []store.Submission{}
	for rows.Next() {
		var sub store.Submission
		var created string
		if err := rows.Scan(&sub.ID, &sub.ClientID, &sub.Client, &sub.FormID, &sub.Form, &sub.FormType, &sub.Status, &sub.Name, &sub.Email, &sub.Subject, &sub.Message, &sub.Priority, &sub.IP, &sub.UserAgent, &created); err != nil {
			return nil, 0, err
		}
		sub.CreatedAt = parseTime(created)
		subs = append(subs, sub)
	}
	return subs, total, rows.Err()
}

func (s *Store) GetSubmission(id int64) (store.Submission, error) {
	row := s.db.QueryRow(`
SELECT s.id, s.client_id, c.name, s.form_id, f.name, f.type, s.status, s.name, s.email, s.subject, s.message, s.priority, s.ip, s.user_agent, s.created_at
FROM submissions s
JOIN clients c ON c.id = s.client_id
JOIN forms f ON f.id = s.form_id
WHERE s.id = ?
`, id)

	var sub store.Submission
	var created string
	if err := row.Scan(&sub.ID, &sub.ClientID, &sub.Client, &sub.FormID, &sub.Form, &sub.FormType, &sub.Status, &sub.Name, &sub.Email, &sub.Subject, &sub.Message, &sub.Priority, &sub.IP, &sub.UserAgent, &created); err != nil {
		return store.Submission{}, err
	}
	sub.CreatedAt = parseTime(created)
	return sub, nil
}

func (s *Store) UpdateSubmissionStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE submissions SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s *Store) DeleteSubmission(id int64) error {
	_, err := s.db.Exec(`DELETE FROM submissions WHERE id = ?`, id)
	return err
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02 15:04:05", value)
	if err == nil {
		return parsed
	}
	parsed, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed
	}
	return time.Time{}
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func formatLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	return limit
}

func formatOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func (s *Store) ListSubmissionsSafe(offset, limit int) ([]store.Submission, int, error) {
	limit = formatLimit(limit)
	offset = formatOffset(offset)
	return s.ListSubmissions(offset, limit)
}

func (s *Store) ListClientsSafe(offset, limit int) ([]store.Client, int, error) {
	limit = formatLimit(limit)
	offset = formatOffset(offset)
	return s.ListClients(offset, limit)
}

func validateFormType(formType store.FormType) error {
	switch formType {
	case store.FormTypeSupport, store.FormTypeContact:
		return nil
	default:
		return fmt.Errorf("invalid form type: %s", formType)
	}
}

func (s *Store) CreateFormSafe(clientID int64, name string, formType store.FormType) (store.Form, error) {
	if err := validateFormType(formType); err != nil {
		return store.Form{}, err
	}
	return s.CreateForm(clientID, name, formType)
}
