package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetTrashedTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		trashedTimestamp, trashedTimestampErr := shared.GetOptionalTimestampQueryParameter(c, "trashedTimestamp")
		id, idErr := shared.GetOptionalInt64QueryParameter(c, "id")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/trashed?trashedTimestamp=%v&id=%v [GET]",
			shared.GetLogValueForOptionalTimestampQueryParameter(trashedTimestamp, trashedTimestampErr),
			shared.GetLogValueForOptionalInt64QueryParameter(id, idErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling get trashed tasks failed"

		if trashedTimestampErr != nil || idErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidCursor := trashedTimestamp == nil && id == nil ||
			trashedTimestamp != nil && id != nil
		if !hasValidCursor {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		taskResponses, err := taskStore.GetTrashedTasks(c.UserId, trashedTimestamp, id)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, taskResponses)
	}
}
