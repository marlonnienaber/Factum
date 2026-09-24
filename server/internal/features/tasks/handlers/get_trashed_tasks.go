package handlers

import (
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetTrashedTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		categoryIds, categoryIdsErr := shared.GetOptionalIdQueryParameterList(c, "categoryId", maximumCategoryFilterCount)
		startDate, startDateErr := shared.GetOptionalDateQueryParameter(c, "startDate")
		endDate, endDateErr := shared.GetOptionalDateQueryParameter(c, "endDate")
		trashedDate, trashedDateErr := shared.GetOptionalTimestampQueryParameter(c, "trashedDate")
		id, idErr := shared.GetOptionalIdQueryParameter(c, "id")

		unexpectedErrorLogMessage := "handling get trashed tasks failed"

		if categoryIdsErr != nil || startDateErr != nil || endDateErr != nil || trashedDateErr != nil || idErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidDateRange := startDate == nil || endDate == nil || !startDate.After(*endDate)
		hasValidCursor := trashedDate == nil && id == nil ||
			trashedDate != nil && id != nil
		if !hasValidCursor || !hasValidDateRange {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		response, err := taskStore.GetTrashedTasks(c.UserId, categoryIds, startDate, endDate, trashedDate, id)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
