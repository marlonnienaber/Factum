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

func NewChangeTaskDateAndOrRankHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		slog.Info("/tasks/change-date-and-or-rank [PATCH]")

		unexpectedErrorLogMessage := "handling change task date and/or rank failed"

		var request transport.ChangeTaskDateAndOrRankRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, validation.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		err := taskService.ChangeTaskDateAndOrRank(c.UserId, request)
		if err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, taskNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskDateAndOrRankChangeNotAllowed):
				c.JSON(http.StatusBadRequest, taskDateAndOrRankChangeNotAllowedErrorResponse)
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

var taskNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_not_found",
	Message:   "No task exists for the provided id.",
}

var taskDateAndOrRankChangeNotAllowedErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_date_and_or_rank_change_not_allowed",
	Message:   "This date and/or rank transition is not allowed.",
}

var taskRankAlreadyExistsErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_rank_already_exists",
	Message:   "The provided rank is already in use.",
}
