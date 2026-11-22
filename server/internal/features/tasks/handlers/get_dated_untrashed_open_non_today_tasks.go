package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetDatedUntrashedOpenNonTodayTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		date, dateErr := shared.GetOptionalDateQueryParameter(c, "date")
		rank, rankErr := shared.GetOptionalRankQueryParameter(c, "rank")

		if dateErr != nil || rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidCursor := date == nil && rank == nil || date != nil && rank != nil
		if !hasValidCursor {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		today := shared.NewDate(time.Now())
		response, err := taskStore.GetDatedUntrashedOpenNonTodayTasks(c.Request.Context(), c.UserId, today, date, rank)
		if err != nil {
			slog.Error("handling get dated untrashed open non-today tasks failed", "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
