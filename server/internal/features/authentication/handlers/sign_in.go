package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewSignInHandler(authenticationService *services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		unexpectedErrorLogMessage := "handling sign-in failed"

		var request transport.SignInRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, shared.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		authenticationResponse, err := authenticationService.SignIn(c.Request.Context(), request, sessions.Default(c))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, invalidCredentialsErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.JSON(http.StatusOK, authenticationResponse)
	}
}

var invalidCredentialsErrorResponse = shared.ErrorResponse{
	ErrorCode: "invalid_credentials",
	Message:   "Email or password incorrect.",
}
