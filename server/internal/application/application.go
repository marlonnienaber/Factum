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

	userStore := stores.NewUserStore(db)
	authenticationService := services.NewAuthenticationService(userStore, cfg)

	app := gin.Default()
	app.Use(shared.NewSessionConfigurationMiddleware(cfg, db))
	api := app.Group("/api")

	api.POST("/sign-up", handlers.NewSignUpHandler(authenticationService))
	api.POST("/sign-in", handlers.NewSignInHandler(authenticationService))
	api.POST("/sign-out", shared.RequireAuth(handlers.NewSignOutHandler(authenticationService)))
	api.GET("/current-authentication", shared.RequireAuth(handlers.NewGetCurrentAuthenticationHandler(userStore)))

	return app
}
