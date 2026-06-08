package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/marlonnienaber/Factum/server/internal/application"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/persistence"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, falling back to development environment.")
	}

	cfg := configuration.LoadConfiguration()

	db := persistence.NewDatabase(cfg.DatabaseUrl)

	app := application.NewApplication(cfg, db)

	defer db.Close()

	if err := app.Run(":" + cfg.Port); err != nil {
		slog.Error("Failed to run server.", "error", err)
		os.Exit(1)
	}
}
