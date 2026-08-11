package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetUndatedTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		rank, rankErr := shared.GetOptionalRankQueryParameter(c, "rank")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/undated?rank=%v [GET]",
			shared.GetLogValueForOptionalStringQueryParameter(rank, rankErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling get undated tasks failed"

		if rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}

		taskResponses, err := taskStore.GetUndatedUntrashedUndoneTasks(c.UserId, rank)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, taskResponses)
	}
}
