package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewSignUpHandler(authenticationService *services.AuthenticationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info("/sign-up [POST]")

		unexpectedErrorLogMessage := "handling sign-up failed"

		var request transport.SignUpRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, validation.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		authenticationResponse, err := authenticationService.SignUp(request, sessions.Default(c))
		if err != nil {
			switch {
			case errors.Is(err, services.ErrEmailNotAllowed):
				c.JSON(http.StatusForbidden, emailNotAllowedErrorResponse)
				return
			case errors.Is(err, stores.ErrEmailAlreadyInUse):
				c.JSON(http.StatusConflict, emailAlreadyInUseErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.JSON(http.StatusCreated, authenticationResponse)
	}
}

var emailNotAllowedErrorResponse = shared.ErrorResponse{
	ErrorCode: "email_not_allowed",
	Message:   "You are not allowed to sign up with this email.",
}

var emailAlreadyInUseErrorResponse = shared.ErrorResponse{
	ErrorCode: "email_already_in_use",
	Message:   "Email is already in use.",
}
