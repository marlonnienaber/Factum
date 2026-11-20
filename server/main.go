package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/marlonnienaber/Factum/server/internal/application"
	"github.com/marlonnienaber/Factum/server/internal/injections"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, falling back to development environment.")
	}

	env := injections.LoadEnvironment()

	db := injections.NewDatabase(env.DatabaseUrl)

	app := application.NewApplication(env, db)

	defer db.Close()

	if err := app.Run(":" + env.ServerPort); err != nil {
		slog.Error("Failed to run server.", "error", err)
		os.Exit(1)
	}
}
