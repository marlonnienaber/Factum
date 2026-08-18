package handlers

import (
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetCategoriesHandler(categoryStore *stores.CategoryStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		slog.Info("/categories [GET]")

		unexpectedErrorLogMessage := "handling get categories failed"

		categoryResponses, err := categoryStore.GetCategories(c.UserId)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, categoryResponses)
	}
}
