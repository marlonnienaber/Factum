package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewSignOutHandler(authenticationService *services.AuthenticationService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		unexpectedErrorLogMessage := "handling sign-out failed"

		session := sessions.Default(c.Context)
		err := authenticationService.DeleteSession(session)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
