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

func NewCreateTaskHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		slog.Info("/tasks [POST]")

		unexpectedErrorLogMessage := "handling create task failed"

		var request transport.CreateTaskRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, validation.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		taskResponse, err := taskService.CreateTask(c.UserId, request)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrCategoryNotFound):
				c.JSON(http.StatusNotFound, categoryNotFoundErrorResponse)
				return
			default:
				slog.Error(unexpectedErrorLogMessage, "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
				return
			}
		}

		c.JSON(http.StatusCreated, taskResponse)
	}
}

var invalidQueryParameterCombinationErrorResponse = shared.ErrorResponse{
	ErrorCode: "invalid_query_parameter_combination",
	Message:   "URL contains an invalid query parameter combination.",
}

var categoryNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "category_not_found",
	Message:   "No category exists for the provided categoryId.",
}
