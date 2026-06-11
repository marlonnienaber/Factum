package support

import (
	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/application"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
)

func TestApp(cfg configuration.Configuration) *gin.Engine {
	return application.NewApplication(cfg, DB)
}
