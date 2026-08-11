package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewRestoreTaskHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		id, idErr := shared.GetRequiredInt64PathVariable(c, "id")
		rank, rankErr := shared.GetRequiredRankQueryParameter(c, "rank")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/%v/restore?rank=%v [PATCH]",
			shared.GetLogValueForRequiredInt64PathVariable(id, idErr),
			shared.GetLogValueForRequiredStringQueryParameter(rank, rankErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling restore task failed"

		if idErr != nil {
			c.JSON(http.StatusBadRequest, invalidTaskIdErrorResponse)
			return
		}
		if errors.Is(rankErr, shared.MissingQueryParameterErr) {
			c.JSON(http.StatusBadRequest, shared.MissingQueryParameterErrorResponse)
			return
		}
		if rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}

		if err := taskService.RestoreTask(c.UserId, id, rank); err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, taskNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskAlreadyRestored):
				c.JSON(http.StatusBadRequest, taskAlreadyRestoredErrorResponse)
				return
			case errors.Is(err, stores.ErrTaskRankAlreadyExists):
				c.JSON(http.StatusConflict, taskRankAlreadyExistsErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.Status(http.StatusNoContent)
	}
}

var taskAlreadyRestoredErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_already_restored",
	Message:   "Task is already restored.",
}
