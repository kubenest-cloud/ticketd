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
	AdminUser     string // Admin dashboard username (required unless DisableAuth is true)
	AdminPass     string // Admin dashboard password (required unless DisableAuth is true)
	PublicBaseURL string // Public base URL for embed scripts (optional, auto-detected if not set)
	CustomCSSPath string // Path to custom CSS file for forms (optional)
	DisableAuth   bool   // Disable built-in authentication (for use with external auth proxies like oauth2-proxy)
}

// Load reads configuration from environment variables.
//
// Required environment variables (unless TICKETD_DISABLE_AUTH=true):
//   - TICKETD_ADMIN_USER: Username for admin dashboard
//   - TICKETD_ADMIN_PASS: Password for admin dashboard
//
// Optional environment variables:
//   - TICKETD_PORT: Server port (default: 8080)
//   - TICKETD_DB_PATH: Database file path (default: ticketd.db)
//   - TICKETD_PUBLIC_BASE_URL: Public URL for production deployments
//   - TICKETD_CUSTOM_CSS: Path to custom CSS file for embedded forms
//   - TICKETD_DISABLE_AUTH: Set to "true" to disable built-in authentication (use with external auth proxies)
func Load() Config {
	cfg := Config{
		Port:          envOrDefault("TICKETD_PORT", "8080"),
		DBPath:        envOrDefault("TICKETD_DB_PATH", "ticketd.db"),
		AdminUser:     strings.TrimSpace(os.Getenv("TICKETD_ADMIN_USER")),
		AdminPass:     os.Getenv("TICKETD_ADMIN_PASS"), // Don't trim password (whitespace might be intentional)
		PublicBaseURL: strings.TrimSpace(os.Getenv("TICKETD_PUBLIC_BASE_URL")),
		CustomCSSPath: strings.TrimSpace(os.Getenv("TICKETD_CUSTOM_CSS")),
		DisableAuth:   strings.ToLower(strings.TrimSpace(os.Getenv("TICKETD_DISABLE_AUTH"))) == "true",
	}
	return cfg
}

// Validate checks that all required configuration is present and valid.
// Returns a descriptive error if any validation fails.
func (c Config) Validate() error {
	// Check required fields (unless auth is disabled)
	if !c.DisableAuth {
		if c.AdminUser == "" {
			return fmt.Errorf("TICKETD_ADMIN_USER is required (or set TICKETD_DISABLE_AUTH=true to use external authentication)")
		}
		if c.AdminPass == "" {
			return fmt.Errorf("TICKETD_ADMIN_PASS is required (or set TICKETD_DISABLE_AUTH=true to use external authentication)")
		}
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
	authStatus := "enabled"
	if c.DisableAuth {
		authStatus = "disabled (using external auth)"
	}
	return fmt.Sprintf("Config{Port: %s, DBPath: %s, Auth: %s, PublicBaseURL: %s, CustomCSSPath: %s}",
		c.Port, c.DBPath, authStatus, c.PublicBaseURL, c.CustomCSSPath)
}

// envOrDefault returns the value of an environment variable or a fallback default.
func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
