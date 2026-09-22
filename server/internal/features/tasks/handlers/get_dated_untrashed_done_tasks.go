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
		date, dateErr := shared.GetOptionalDateQueryParameter(c, "date")
		rank, rankErr := shared.GetOptionalRankQueryParameter(c, "rank")

		if categoryIdsErr != nil || dateErr != nil || rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidCursor := date == nil && rank == nil || date != nil && rank != nil
		if !hasValidCursor {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		response, err := taskStore.GetDatedUntrashedDoneTasks(c.UserId, categoryIds, date, rank)
		if err != nil {
			slog.Error("handling get dated untrashed done tasks failed", "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
