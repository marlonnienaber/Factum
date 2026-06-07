package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewSignOutHandler(authenticationService *services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info("/sign-out [POST]")

		unexpectedErrorLogMessage := "handling sign-out failed"

		session := sessions.Default(c)
		err := authenticationService.DeleteSession(session)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
