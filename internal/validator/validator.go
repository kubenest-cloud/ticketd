// Package validator provides input validation functions for TicketD.
// It validates user input, form data, and other inputs to ensure data integrity
// and prevent invalid data from entering the system.
package validator

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	"ticketd/internal/errors"
	"ticketd/internal/store"
)

const (
	// Field length constraints
	minNameLength    = 1
	maxNameLength    = 255
	minDomainLength  = 3
	maxDomainLength  = 255
	minEmailLength   = 3
	maxEmailLength   = 255
	minSubjectLength = 1
	maxSubjectLength = 500
	minMessageLength = 1
	maxMessageLength = 10000
	maxPriorityLength = 50
)

// Status constants for submission status validation
const (
	StatusOpen       = "OPEN"
	StatusInProgress = "IN_PROGRESS"
	StatusClosed     = "CLOSED"
)

// ValidateFormType checks if the provided form type is valid.
// Valid types are "support" and "contact".
func ValidateFormType(formType store.FormType) error {
	switch formType {
	case store.FormTypeSupport, store.FormTypeContact:
		return nil
	default:
		return errors.InvalidInputError("form type", fmt.Sprintf("must be %q or %q", store.FormTypeSupport, store.FormTypeContact))
	}
}

// ValidateStatus checks if the provided status is valid.
// Valid statuses are OPEN, IN_PROGRESS, and CLOSED.
func ValidateStatus(status string) error {
	switch status {
	case StatusOpen, StatusInProgress, StatusClosed:
		return nil
	default:
		return errors.InvalidInputError("status", fmt.Sprintf("must be %q, %q, or %q", StatusOpen, StatusInProgress, StatusClosed))
	}
}

// ValidateEmail checks if the provided email address is valid.
func ValidateEmail(email string) error {
	if email == "" {
		// Email is optional in some contexts
		return nil
	}

	if len(email) < minEmailLength {
		return errors.InvalidInputError("email", fmt.Sprintf("must be at least %d characters", minEmailLength))
	}

	if len(email) > maxEmailLength {
		return errors.InvalidInputError("email", fmt.Sprintf("must be at most %d characters", maxEmailLength))
	}

	// Use standard library to validate email format
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.InvalidInputError("email", "invalid email format")
	}

	return nil
}

// ValidateName validates a name field (client name, form name, etc.).
func ValidateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.InvalidInputError("name", "cannot be empty")
	}

	if len(name) < minNameLength {
		return errors.InvalidInputError("name", fmt.Sprintf("must be at least %d character", minNameLength))
	}

	if len(name) > maxNameLength {
		return errors.InvalidInputError("name", fmt.Sprintf("must be at most %d characters", maxNameLength))
	}

	return nil
}

// ValidateDomain validates a domain name for allowed_domain field.
// It checks format and length constraints.
func ValidateDomain(domain string) error {
	domain = strings.TrimSpace(domain)

	if domain == "" {
		return errors.InvalidInputError("domain", "cannot be empty")
	}

	if len(domain) < minDomainLength {
		return errors.InvalidInputError("domain", fmt.Sprintf("must be at least %d characters", minDomainLength))
	}

	if len(domain) > maxDomainLength {
		return errors.InvalidInputError("domain", fmt.Sprintf("must be at most %d characters", maxDomainLength))
	}

	// Basic domain format validation
	// Accept both domain.com and https://domain.com formats
	testURL := domain
	if !strings.Contains(domain, "://") {
		testURL = "https://" + domain
	}

	parsedURL, err := url.Parse(testURL)
	if err != nil || parsedURL.Host == "" {
		return errors.InvalidInputError("domain", "invalid domain format")
	}

	return nil
}

// ValidateString validates a general string field with min and max length constraints.
func ValidateString(fieldName, value string, minLength, maxLength int, required bool) error {
	value = strings.TrimSpace(value)

	if value == "" && required {
		return errors.InvalidInputError(fieldName, "is required")
	}

	if value == "" && !required {
		// Optional field, empty is acceptable
		return nil
	}

	if len(value) < minLength {
		return errors.InvalidInputError(fieldName, fmt.Sprintf("must be at least %d characters", minLength))
	}

	if len(value) > maxLength {
		return errors.InvalidInputError(fieldName, fmt.Sprintf("must be at most %d characters", maxLength))
	}

	return nil
}

// ValidateClient validates client creation/update input.
func ValidateClient(name, allowedDomain string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	if err := ValidateDomain(allowedDomain); err != nil {
		return err
	}

	return nil
}

// ValidateForm validates form creation input.
func ValidateForm(name string, formType store.FormType) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	if err := ValidateFormType(formType); err != nil {
		return err
	}

	return nil
}

// ValidateSubmission validates submission input before storing in database.
func ValidateSubmission(input store.SubmissionInput) error {
	// Name is optional for some form types
	if input.Name != "" {
		if err := ValidateString("name", input.Name, minNameLength, maxNameLength, false); err != nil {
			return err
		}
	}

	// Email validation (optional field)
	if err := ValidateEmail(input.Email); err != nil {
		return err
	}

	// Subject validation (optional field)
	if input.Subject != "" {
		if err := ValidateString("subject", input.Subject, minSubjectLength, maxSubjectLength, false); err != nil {
			return err
		}
	}

	// Message is required
	if err := ValidateString("message", input.Message, minMessageLength, maxMessageLength, true); err != nil {
		return err
	}

	// Priority is optional
	if input.Priority != "" {
		if err := ValidateString("priority", input.Priority, 1, maxPriorityLength, false); err != nil {
			return err
		}
	}

	return nil
}

// TrimAndValidateClient trims whitespace and validates client input.
// Returns the trimmed values and any validation error.
func TrimAndValidateClient(name, allowedDomain string) (string, string, error) {
	name = strings.TrimSpace(name)
	allowedDomain = strings.TrimSpace(allowedDomain)

	if err := ValidateClient(name, allowedDomain); err != nil {
		return "", "", err
	}

	return name, allowedDomain, nil
}

// TrimSubmissionInput trims whitespace from all string fields in submission input.
func TrimSubmissionInput(input store.SubmissionInput) store.SubmissionInput {
	return store.SubmissionInput{
		Name:      strings.TrimSpace(input.Name),
		Email:     strings.TrimSpace(input.Email),
		Subject:   strings.TrimSpace(input.Subject),
		Message:   strings.TrimSpace(input.Message),
		Priority:  strings.TrimSpace(input.Priority),
		IP:        strings.TrimSpace(input.IP),
		UserAgent: strings.TrimSpace(input.UserAgent),
	}
}
