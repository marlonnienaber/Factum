package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewBatchTrashTasksHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		trashedDate, trashedDateErr := shared.GetRequiredTimestampQueryParameter(c, "trashedDate")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/batch-trash?trashedDate=%v [PATCH]",
			shared.GetLogValueForRequiredTimestampQueryParameter(trashedDate, trashedDateErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling batch trash tasks failed"

		if errors.Is(trashedDateErr, shared.MissingQueryParameterErr) {
			c.JSON(http.StatusBadRequest, shared.MissingQueryParameterErrorResponse)
			return
		}
		if trashedDateErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}

		var request transport.BatchTrashTasksRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, validation.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		if err := taskService.BatchTrashTasks(c.UserId, request.TaskIds, trashedDate); err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, oneOfProvidedTasksNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskAlreadyTrashed):
				c.JSON(http.StatusBadRequest, oneOfProvidedTasksAlreadyTrashedErrorResponse)
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

var oneOfProvidedTasksAlreadyTrashedErrorResponse = shared.ErrorResponse{
	ErrorCode: "one_of_provided_tasks_already_trashed",
	Message:   "One of the provided tasks is already trashed.",
}
