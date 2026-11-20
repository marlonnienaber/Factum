package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewSignUpHandler(authenticationService *services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		unexpectedErrorLogMessage := "handling sign-up failed"

		var request transport.SignUpRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, shared.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		err := authenticationService.SignUp(request)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrSignUpNotAvailable):
				c.JSON(http.StatusBadRequest, signUpNotAvailableErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.Status(http.StatusCreated)
	}
}

var signUpNotAvailableErrorResponse = shared.ErrorResponse{
	ErrorCode: "sign_up_not_available",
	Message:   "Sign-up is not available. Please restart the sign-up process.",
}
