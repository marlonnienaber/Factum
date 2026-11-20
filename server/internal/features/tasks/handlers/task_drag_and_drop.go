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

func NewTaskDragAndDropHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		taskId, taskIdErr := shared.GetRequiredInt64PathVariable(c, "id")
		if taskIdErr != nil {
			c.JSON(http.StatusBadRequest, invalidTaskIdErrorResponse)
			return
		}

		var request transport.TaskDragAndDropRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				c.JSON(http.StatusBadRequest, shared.NewValidationErrorResponse(validationErrors))
				return
			}
			c.JSON(http.StatusBadRequest, shared.JsonParsingErrorResponse)
			return
		}

		response, err := taskService.DragAndDropTask(c.UserId, taskId, request)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrPreviousTaskNotFound):
				c.JSON(http.StatusNotFound, previousTaskNotFoundErrorResponse)
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, taskNotFoundErrorResponse)
			case errors.Is(err, services.ErrRequestDataRepresentInvalidDragAndDropGesture):
				c.JSON(http.StatusBadRequest, invalidDragAndDropGestureErrorResponse)
			default:
				slog.Error("handling task drag and drop failed", "error", err)
				c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			}
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

var taskNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "task_not_found",
	Message:   "No task exists for the provided id.",
}

var previousTaskNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "previous_task_not_found",
	Message:   "No previous task exists for the provided previousTaskId.",
}

var invalidDragAndDropGestureErrorResponse = shared.ErrorResponse{
	ErrorCode: "invalid_drag_and_drop_gesture",
	Message:   "The combination of task ID and request body properties does not represent a valid drag and drop gesture.",
}
