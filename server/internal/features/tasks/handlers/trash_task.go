package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewTrashTaskHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		id, idErr := shared.GetRequiredInt64PathVariable(c, "id")

		unexpectedErrorLogMessage := "handling trash task failed"

		if idErr != nil {
			c.JSON(http.StatusBadRequest, invalidTaskIdErrorResponse)
			return
		}
		response, err := taskService.TrashTask(c.UserId, id)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, taskNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskAlreadyTrashed):
				c.JSON(http.StatusBadRequest, taskAlreadyTrashedErrorResponse)
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

var invalidTaskIdErrorResponse = shared.ErrorResponse{
	ErrorCode: "invalid_task_id",
	Message:   "Task id must be a positive integer.",
}

var taskAlreadyTrashedErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_already_trashed",
	Message:   "Task is already trashed.",
}
