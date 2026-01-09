package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"ticketd/internal/config"
	"ticketd/internal/store/sqlite"
	"ticketd/internal/web"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting TicketD")

	// Load .env file if present (development only)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			slog.Error("Failed to load .env file", "error", err)
			os.Exit(1)
		}
		slog.Info("Loaded configuration from .env file")
	}

	// Load and validate configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Configuration loaded successfully", "config", cfg.String())

	// Initialize database
	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err, "db_path", cfg.DBPath)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()
	slog.Info("Database initialized", "db_path", cfg.DBPath)

	// Run database migrations
	if err := store.Migrate(); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Database migrations completed")

	// Initialize web application
	app, err := web.NewApp(cfg, store)
	if err != nil {
		slog.Error("Failed to initialize web application", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	addr := ":" + cfg.Port
	slog.Info("Starting HTTP server", "address", addr)
	if err := http.ListenAndServe(addr, app.Router()); err != nil {
		slog.Error("HTTP server failed", "error", err, "address", addr)
		os.Exit(1)
	}
}
