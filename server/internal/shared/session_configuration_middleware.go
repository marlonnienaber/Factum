package shared

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
)

func NewSessionConfigurationMiddleware(cfg configuration.Configuration, db *sqlx.DB) gin.HandlerFunc {
	store, err := postgres.NewStore(db.DB, []byte(cfg.SessionSecretKey))
	if err != nil {
		slog.Error("Failed to initialize session middleware.", "error", err)
		os.Exit(1)
	}

	store.Options(sessions.Options{
		Path:     "/",
		Secure:   cfg.SessionCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return sessions.Sessions("factumSession", store)
}
