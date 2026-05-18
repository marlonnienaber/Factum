package support

import (
	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/application"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
)

func TestApp() *gin.Engine {
	cfg := configuration.Configuration{
		Port: ":3001",
	}

	return application.NewApplication(cfg, DB)
}
