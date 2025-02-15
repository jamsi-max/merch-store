package main

import (
	"log"
	"net/http"

	"github.com/jamsi-max/merch-store/config"
	"github.com/jamsi-max/merch-store/internal/db"
	"github.com/jamsi-max/merch-store/internal/router"
)

const (
	red   = "\033[31m"
	reset = "\033[0m"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf(red+"[ERR]"+reset+"failed to load config: %v", err)
	}

	db, err := db.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf(red+"[ERR]"+reset+"failed to connect to database: %v", err)
	}
	defer db.DB.Close()

	server := router.SetupRouter(db, cfg.JWTSecret)

	if err := server.Run(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatalf(red+"[ERR]"+reset+" failed to start server: %v", err)
	}
}
