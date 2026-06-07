package application

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/handlers"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewApplication(cfg configuration.Configuration, db *sqlx.DB) *gin.Engine {
	validation.RegisterJsonPropertyNames()

	app := gin.Default()

	app.Use(shared.NewSessionConfigurationMiddleware(cfg, db))

	userStore := stores.NewUserStore(db)
	authenticationService := services.NewAuthenticationService(userStore, cfg)

	app.POST("/sign-up", handlers.NewSignUpHandler(authenticationService))
	app.POST("/sign-in", handlers.NewSignInHandler(authenticationService))
	app.POST("/sign-out", shared.RequireAuth(handlers.NewSignOutHandler(authenticationService)))

	return app
}
