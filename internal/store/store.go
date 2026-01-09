package store

import "time"

type Client struct {
	ID            int64
	Name          string
	AllowedDomain string
	CreatedAt     time.Time
}

type FormType string

const (
	FormTypeSupport FormType = "support"
	FormTypeContact FormType = "contact"
)

type Form struct {
	ID        int64
	ClientID  int64
	Name      string
	Type      FormType
	CreatedAt time.Time
}

type Submission struct {
	ID        int64
	ClientID  int64
	Client    string
	FormID    int64
	Form      string
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

type SubmissionInput struct {
	Name      string
	Email     string
	Subject   string
	Message   string
	Priority  string
	IP        string
	UserAgent string
}

type Store interface {
	Migrate() error
	Close() error

	CreateClient(name, allowedDomain string) (Client, error)
	ListClients(offset, limit int) ([]Client, int, error)
	GetClient(id int64) (Client, error)
	UpdateClient(id int64, name, allowedDomain string) error

	CreateForm(clientID int64, name string, formType FormType) (Form, error)
	ListForms(clientID int64) ([]Form, error)
	GetForm(id int64) (Form, error)

	CreateSubmission(formID int64, input SubmissionInput) (Submission, error)
	ListSubmissions(offset, limit int) ([]Submission, int, error)
	GetSubmission(id int64) (Submission, error)
	UpdateSubmissionStatus(id int64, status string) error
	DeleteSubmission(id int64) error
}
