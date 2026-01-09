// Package config provides configuration loading from environment variables.
// It supports .env files (via godotenv) and provides sensible defaults for all optional settings.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values for TicketD.
// Values are loaded from environment variables with sensible defaults where appropriate.
type Config struct {
	Port          string // Server port (default: 8080)
	DBPath        string // SQLite database file path (default: ticketd.db)
	AdminUser     string // Admin dashboard username (required)
	AdminPass     string // Admin dashboard password (required)
	PublicBaseURL string // Public base URL for embed scripts (optional, auto-detected if not set)
	CustomCSSPath string // Path to custom CSS file for forms (optional)
}

// Load reads configuration from environment variables.
//
// Required environment variables:
//   - TICKETD_ADMIN_USER: Username for admin dashboard
//   - TICKETD_ADMIN_PASS: Password for admin dashboard
//
// Optional environment variables:
//   - TICKETD_PORT: Server port (default: 8080)
//   - TICKETD_DB_PATH: Database file path (default: ticketd.db)
//   - TICKETD_PUBLIC_BASE_URL: Public URL for production deployments
//   - TICKETD_CUSTOM_CSS: Path to custom CSS file for embedded forms
func Load() Config {
	cfg := Config{
		Port:          envOrDefault("TICKETD_PORT", "8080"),
		DBPath:        envOrDefault("TICKETD_DB_PATH", "ticketd.db"),
		AdminUser:     strings.TrimSpace(os.Getenv("TICKETD_ADMIN_USER")),
		AdminPass:     os.Getenv("TICKETD_ADMIN_PASS"), // Don't trim password (whitespace might be intentional)
		PublicBaseURL: strings.TrimSpace(os.Getenv("TICKETD_PUBLIC_BASE_URL")),
		CustomCSSPath: strings.TrimSpace(os.Getenv("TICKETD_CUSTOM_CSS")),
	}
	return cfg
}

// Validate checks that all required configuration is present and valid.
// Returns a descriptive error if any validation fails.
func (c Config) Validate() error {
	// Check required fields
	if c.AdminUser == "" {
		return fmt.Errorf("TICKETD_ADMIN_USER is required")
	}
	if c.AdminPass == "" {
		return fmt.Errorf("TICKETD_ADMIN_PASS is required")
	}

	// Validate port number
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid TICKETD_PORT %q: must be a number", c.Port)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid TICKETD_PORT %d: must be between 1 and 65535", port)
	}

	// Validate DB path is not empty
	if c.DBPath == "" {
		return fmt.Errorf("TICKETD_DB_PATH cannot be empty")
	}

	// Validate custom CSS path exists if specified
	if c.CustomCSSPath != "" {
		if _, err := os.Stat(c.CustomCSSPath); err != nil {
			return fmt.Errorf("TICKETD_CUSTOM_CSS file %q not found or not accessible: %w", c.CustomCSSPath, err)
		}
	}

	return nil
}

// String returns a string representation of the config with sensitive values redacted.
// Useful for logging configuration at startup.
func (c Config) String() string {
	return fmt.Sprintf("Config{Port: %s, DBPath: %s, AdminUser: %s, AdminPass: *****, PublicBaseURL: %s, CustomCSSPath: %s}",
		c.Port, c.DBPath, c.AdminUser, c.PublicBaseURL, c.CustomCSSPath)
}

// envOrDefault returns the value of an environment variable or a fallback default.
func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
