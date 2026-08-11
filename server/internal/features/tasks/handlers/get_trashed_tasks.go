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
		trashedDate, trashedDateErr := shared.GetOptionalTimestampQueryParameter(c, "trashedDate")
		id, idErr := shared.GetOptionalInt64QueryParameter(c, "id")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/trashed?trashedDate=%v&id=%v [GET]",
			shared.GetLogValueForOptionalTimestampQueryParameter(trashedDate, trashedDateErr),
			shared.GetLogValueForOptionalInt64QueryParameter(id, idErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling get trashed tasks failed"

		if trashedDateErr != nil || idErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidCursor := trashedDate == nil && id == nil ||
			trashedDate != nil && id != nil
		if !hasValidCursor {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		taskResponses, err := taskStore.GetTrashedTasks(c.UserId, trashedDate, id)
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, taskResponses)
	}
}
