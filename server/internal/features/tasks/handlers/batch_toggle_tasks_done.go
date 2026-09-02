package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewBatchToggleTaskDoneHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		slog.Info("/tasks/batch-toggle-done [PATCH]")

		unexpectedErrorLogMessage := "handling batch toggle task done failed"

		var request transport.BatchToggleTaskDoneRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, validation.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		response, err := taskService.BatchToggleTaskDone(c.UserId, request.TaskIds)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, oneOfProvidedTasksNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTasksHaveDifferentDoneStates):
				c.JSON(http.StatusBadRequest, tasksHaveDifferentDoneStatesErrorResponse)
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

var oneOfProvidedTasksNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "one_of_provided_tasks_not_found",
	Message:   "One of the provided tasks could not be found.",
}

var tasksHaveDifferentDoneStatesErrorResponse = shared.ErrorResponse{
	ErrorCode: "tasks_have_different_done_states",
	Message:   "All tasks in a batch must have the same done state.",
}
