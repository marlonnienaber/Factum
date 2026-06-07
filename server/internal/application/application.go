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

	userStore := stores.NewUserStore(db)
	authenticationService := services.NewAuthenticationService(userStore, cfg)

	app.Use(shared.NewSessionMiddleware(cfg, db))

	app.POST("/sign-up", handlers.NewSignUpHandler(authenticationService))

	return app
}
