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

		handlerEntryLog := fmt.Sprintf("/tasks/%v/restore [PATCH]", shared.GetLogValueForRequiredInt64PathVariable(id, idErr))
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling restore task failed"

		if idErr != nil {
			c.JSON(http.StatusBadRequest, invalidTaskIdErrorResponse)
			return
		}
		response, err := taskService.RestoreTask(c.UserId, id)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, taskNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskAlreadyRestored):
				c.JSON(http.StatusBadRequest, taskAlreadyRestoredErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

var taskAlreadyRestoredErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_already_restored",
	Message:   "Task is already restored.",
}
