package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewDeleteCategoryHandler(categoryStore *stores.CategoryStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		id, idErr := shared.GetRequiredInt64PathVariable(c, "id")

		unexpectedErrorLogMessage := "handling delete category failed"

		if idErr != nil {
			c.JSON(http.StatusBadRequest, invalidCategoryIdErrorResponse)
			return
		}

		err := categoryStore.DeleteCategory(c.Request.Context(), id, c.UserId)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrCategoryNotFound):
				c.JSON(http.StatusNotFound, categoryNotFoundErrorResponse)
				return
			case errors.Is(err, stores.ErrCategoryStillReferenced):
				c.JSON(http.StatusConflict, categoryStillReferencedErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.Status(http.StatusNoContent)
	}
}

var invalidCategoryIdErrorResponse = shared.ErrorResponse{
	ErrorCode: "invalid_category_id",
	Message:   "Category id must be a positive integer.",
}

var categoryNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "category_not_found",
	Message:   "No category of the logged in user exists with the provided id.",
}

var categoryStillReferencedErrorResponse = shared.ErrorResponse{
	ErrorCode: "category_still_referenced",
	Message:   "Category cannot be deleted while it is referenced by tasks.",
}
