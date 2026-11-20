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
)

func NewBatchCompleteTasksHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		unexpectedErrorLogMessage := "handling batch complete tasks failed"

		var request transport.BatchUpdateDoneRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, shared.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		response, err := taskService.BatchCompleteTasks(c.UserId, request.TaskIds)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, oneOfProvidedTasksNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskAlreadyCompleted):
				c.JSON(http.StatusBadRequest, oneOfProvidedTasksAlreadyCompletedErrorResponse)
				return
			case errors.Is(err, services.ErrTaskIsTrashed):
				c.JSON(http.StatusBadRequest, oneOfProvidedTasksIsTrashedErrorResponse)
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

var oneOfProvidedTasksAlreadyCompletedErrorResponse = shared.ErrorResponse{
	ErrorCode: "one_of_provided_tasks_already_completed",
	Message:   "One of the provided tasks is already completed.",
}
