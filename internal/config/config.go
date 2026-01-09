package config

import (
	"os"
)

type Config struct {
	Port          string
	DBPath        string
	AdminUser     string
	AdminPass     string
	PublicBaseURL string
	CustomCSSPath string
}

func Load() Config {
	cfg := Config{
		Port:          envOrDefault("TICKETD_PORT", "8080"),
		DBPath:        envOrDefault("TICKETD_DB_PATH", "ticketd.db"),
		AdminUser:     os.Getenv("TICKETD_ADMIN_USER"),
		AdminPass:     os.Getenv("TICKETD_ADMIN_PASS"),
		PublicBaseURL: os.Getenv("TICKETD_PUBLIC_BASE_URL"),
		CustomCSSPath: os.Getenv("TICKETD_CUSTOM_CSS"),
	}
	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
