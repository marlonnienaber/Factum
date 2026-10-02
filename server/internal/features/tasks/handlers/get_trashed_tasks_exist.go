package handlers

import (
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetTrashedTasksExistHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		trashedTasksExist, err := taskStore.DoTrashedTasksExist(c.UserId)
		if err != nil {
			slog.Error("handling get whether trashed tasks exist failed", "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, transport.GetTrashedTasksExistResponse{TrashedTasksExist: trashedTasksExist})
	}
}
