package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewCreateCategoryHandler(categoryStore *stores.CategoryStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		title, titleErr := shared.GetRequiredStringQueryParameterWithLength(
			c,
			"title",
			minimumCategoryTitleLength,
			maximumCategoryTitleLength,
		)

		handlerEntryLog := fmt.Sprintf(
			"/categories?title=%v [POST]",
			shared.GetLogValueForRequiredStringQueryParameter(title, titleErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling create category failed"

		if errors.Is(titleErr, shared.MissingQueryParameterErr) {
			c.JSON(http.StatusBadRequest, shared.MissingQueryParameterErrorResponse)
			return
		}
		if titleErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}

		categoryResponse, err := categoryStore.InsertCategory(c.UserId, title)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrCategoryTitleAlreadyInUse):
				c.JSON(http.StatusConflict, categoryTitleAlreadyInUseErrorResponse)
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

const (
	minimumCategoryTitleLength = 1
	maximumCategoryTitleLength = 50
)

var categoryTitleAlreadyInUseErrorResponse = shared.ErrorResponse{
	ErrorCode: "category_title_already_in_use",
	Message:   "Category title is already in use.",
}
