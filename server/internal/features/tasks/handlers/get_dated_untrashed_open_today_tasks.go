package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetDatedUntrashedOpenTodayTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		rank, rankErr := shared.GetOptionalRankQueryParameter(c, "rank")

		if rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}

		today := shared.NewDate(time.Now())
		response, err := taskStore.GetDatedUntrashedOpenTodayTasks(c.UserId, today, rank)
		if err != nil {
			slog.Error("handling get dated untrashed open today tasks failed", "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
