package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewCreateCategoryHandler(categoryStore *stores.CategoryStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		slog.Info("/categories [POST]")

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

		categoryResponse, err := categoryStore.InsertCategory(c.UserId, request.Title, request.Color)
		if err != nil {
			switch {
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
