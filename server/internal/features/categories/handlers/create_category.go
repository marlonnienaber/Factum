package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/services"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewCreateCategoryHandler(categoryService *services.CategoryService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		unexpectedErrorLogMessage := "handling create category failed"

		var request transport.CreateCategoryRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, validation.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		categoryResponse, err := categoryService.CreateCategory(c.UserId, request)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrMaximumCategoryCountReached):
				c.JSON(http.StatusConflict, maximumCategoryCountReachedErrorResponse)
				return
			case errors.Is(err, stores.ErrCategoryTitleAlreadyInUse):
				c.JSON(http.StatusConflict, categoryTitleAlreadyInUseErrorResponse)
				return
			case errors.Is(err, stores.ErrCategoryColorAlreadyInUse):
				c.JSON(http.StatusConflict, categoryColorAlreadyInUseErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.JSON(http.StatusCreated, categoryResponse)
	}
}

var categoryTitleAlreadyInUseErrorResponse = shared.ErrorResponse{
	ErrorCode: "category_title_already_in_use",
	Message:   "Category title is already in use.",
}

var categoryColorAlreadyInUseErrorResponse = shared.ErrorResponse{
	ErrorCode: "category_color_already_in_use",
	Message:   "Category color is already in use.",
}

var maximumCategoryCountReachedErrorResponse = shared.ErrorResponse{
	ErrorCode: "maximum_category_count_reached",
	Message:   "A maximum of 30 categories can be created per user.",
}
