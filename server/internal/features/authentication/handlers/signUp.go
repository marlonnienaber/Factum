package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewSignUpHandler(cfg configuration.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// TODO: check if email is whitelisted
		// TODO: try to register user
		// TODO: handle errors

		c.Status(http.StatusNoContent)
	}
}
