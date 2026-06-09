package handlers

import (
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetCurrentAuthenticationHandler(userStore *stores.UserStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		slog.Info("/current-authentication [GET]")

		unexpectedErrorLogMessage := "handling get current authentication failed"

		doesUserExist, err := userStore.DoesUserExistForId(c.UserId)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}
		if !doesUserExist {
			c.JSON(http.StatusUnauthorized, noUserForSessionErrorResponse)
			return
		}

		authenticationResponse := transport.AuthenticationResponse{UserId: c.UserId}
		c.JSON(http.StatusOK, authenticationResponse)
	}
}

var noUserForSessionErrorResponse = shared.ErrorResponse{
	ErrorCode: "no_user_for_session",
	Message:   "The user account for this session was deleted. Please sign-in with a different account.",
}
