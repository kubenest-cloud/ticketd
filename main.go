package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"ticketd/internal/config"
	"ticketd/internal/store/sqlite"
	"ticketd/internal/web"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}
	cfg := config.Load()
	if cfg.AdminUser == "" || cfg.AdminPass == "" {
		log.Fatal("TICKETD_ADMIN_USER and TICKETD_ADMIN_PASS must be set")
	}

	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatal(err)
	}

	app, err := web.NewApp(cfg, store)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + cfg.Port
	log.Printf("TicketD listening on %s", addr)
	if err := http.ListenAndServe(addr, app.Router()); err != nil {
		log.Fatal(err)
	}
}
