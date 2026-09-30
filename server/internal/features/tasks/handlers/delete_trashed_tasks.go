package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewDeleteTrashedTasksHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		unexpectedErrorLogMessage := "handling delete trashed tasks failed"

		if err := taskStore.DeleteTrashedTasks(c.UserId); err != nil {
			if errors.Is(err, stores.ErrNoTrashedTasks) {
				c.JSON(http.StatusBadRequest, noTrashedTasksToDeleteErrorResponse)
				return
			}
			slog.Error(unexpectedErrorLogMessage, "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

var noTrashedTasksToDeleteErrorResponse = shared.ErrorResponse{
	ErrorCode: "no_trashed_tasks_to_delete",
	Message:   "No trashed tasks exist that could be deleted.",
}
