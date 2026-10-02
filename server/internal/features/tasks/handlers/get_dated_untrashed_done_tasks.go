package handlers

import (
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

const maximumCategoryFilterCount = 30

func NewGetDatedUntrashedDoneTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		categoryIds, categoryIdsErr := shared.GetOptionalIdQueryParameterList(c, "categoryId", maximumCategoryFilterCount)
		startDate, startDateErr := shared.GetOptionalDateQueryParameter(c, "startDate")
		endDate, endDateErr := shared.GetOptionalDateQueryParameter(c, "endDate")
		cursorDate, cursorDateErr := shared.GetOptionalDateQueryParameter(c, "cursorDate")
		cursorRank, cursorRankErr := shared.GetOptionalRankQueryParameter(c, "cursorRank")

		if categoryIdsErr != nil || startDateErr != nil || endDateErr != nil || cursorDateErr != nil || cursorRankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidDateRange := startDate == nil || endDate == nil || !startDate.After(*endDate)
		hasValidCursor := cursorDate == nil && cursorRank == nil || cursorDate != nil && cursorRank != nil
		if !hasValidCursor || !hasValidDateRange {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		response, err := taskStore.GetDatedUntrashedDoneTasks(c.UserId, categoryIds, startDate, endDate, cursorDate, cursorRank)
		if err != nil {
			slog.Error("handling get dated untrashed done tasks failed", "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
