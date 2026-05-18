package application

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
)

func NewApplication(cfg configuration.Configuration, db *sqlx.DB) *gin.Engine {
	app := gin.Default()

	app.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	return app
}
