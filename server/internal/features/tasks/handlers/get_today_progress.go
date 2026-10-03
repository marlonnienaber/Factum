package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewGetTodayProgressHandler(taskStore *stores.TaskStore) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		today := shared.NewDate(time.Now())
		response, err := taskStore.GetTodayProgress(c.UserId, today)
		if err != nil {
			slog.Error("handling get today's task progress failed", "error", err)
			c.JSON(http.StatusInternalServerError, shared.UnexpectedErrorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
