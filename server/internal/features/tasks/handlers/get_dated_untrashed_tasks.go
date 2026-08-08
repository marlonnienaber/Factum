package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetDatedUntrashedTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		open, openErr := shared.GetRequiredBooleanQueryParameter(c, "open")
		date, dateErr := shared.GetOptionalDateQueryParameter(c, "date")
		rank := shared.GetOptionalStringQueryParameter(c, "rank")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/dated/untrashed?open=%v&date=%v&rank=%v [GET]",
			shared.GetLogValueForRequiredBoolQueryParameter(open, openErr),
			shared.GetLogValueForOptionalDateQueryParameter(date, dateErr),
			shared.GetLogValueForOptionalStringQueryParameter(rank),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling get dated untrashed tasks failed"

		if errors.Is(openErr, shared.MissingQueryParameterErr) {
			c.JSON(http.StatusBadRequest, shared.MissingQueryParameterErrorResponse)
			return
		}
		if openErr != nil || dateErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}
		hasValidCursor := date == nil && rank == nil ||
			date != nil && rank != nil
		if !hasValidCursor {
			c.JSON(http.StatusBadRequest, invalidQueryParameterCombinationErrorResponse)
			return
		}

		var taskResponses []transport.TaskResponse
		var err error
		if open {
			taskResponses, err = taskStore.GetDatedUntrashedUndoneTasks(c.UserId, date, rank)
		} else {
			taskResponses, err = taskStore.GetDatedUntrashedDoneTasks(c.UserId, date, rank)
		}
		if err != nil {
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, taskResponses)
	}
}
