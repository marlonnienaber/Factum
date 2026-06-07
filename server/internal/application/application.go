package application

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewApplication(cfg configuration.Configuration, db *sqlx.DB) *gin.Engine {
	validation.RegisterJsonPropertyNames()

	app := gin.Default()

	app.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	return app
}
