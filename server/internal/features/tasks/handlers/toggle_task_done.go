package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

func NewToggleTaskDoneHandler(taskService *services.TaskService) shared.AuthenticatedHandlerFunc {
	return func(c *shared.AuthenticatedContext) {
		id, idErr := shared.GetRequiredInt64PathVariable(c, "id")
		rank, rankErr := shared.GetOptionalRankQueryParameter(c, "rank")

		handlerEntryLog := fmt.Sprintf(
			"/tasks/%v/toggle-done?rank=%v [PATCH]",
			shared.GetLogValueForRequiredInt64PathVariable(id, idErr),
			shared.GetLogValueForOptionalStringQueryParameter(rank, rankErr),
		)
		slog.Info(handlerEntryLog)

		unexpectedErrorLogMessage := "handling toggle task done failed"

		if idErr != nil {
			c.JSON(http.StatusBadRequest, invalidTaskIdErrorResponse)
			return
		}
		if rankErr != nil {
			c.JSON(http.StatusBadRequest, shared.InvalidQueryParameterErrorResponse)
			return
		}

		if err := taskService.ToggleTaskDone(c.UserId, id, rank); err != nil {
			switch {
			case errors.Is(err, stores.ErrTaskNotFound):
				c.JSON(http.StatusNotFound, taskNotFoundErrorResponse)
				return
			case errors.Is(err, services.ErrTaskRankRequired):
				c.JSON(http.StatusBadRequest, shared.MissingQueryParameterErrorResponse)
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
