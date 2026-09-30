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
		rank, rankErr := shared.GetOptionalRankQueryParameter(c, "rank")

		unexpectedErrorLogMessage := "handling get trashed tasks failed"

		if categoryIdsErr != nil || startDateErr != nil || endDateErr != nil || rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidDateRange := startDate == nil || endDate == nil || !startDate.After(*endDate)
		if !hasValidDateRange {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		response, err := taskStore.GetTrashedTasks(c.UserId, categoryIds, startDate, endDate, rank)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
