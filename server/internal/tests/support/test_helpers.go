package support

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func PrepareTestEnvironment(t *testing.T) *gin.Engine {
	t.Helper()
	databaseReset(t)
	return TestApp()
}
