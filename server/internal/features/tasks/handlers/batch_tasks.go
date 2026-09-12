package handlers

import "github.com/marlonnienaber/Factum/server/internal/shared"

var oneOfProvidedTasksNotFoundErrorResponse = shared.ErrorResponse{
	ErrorCode: "one_of_provided_tasks_not_found",
	Message:   "One of the provided tasks could not be found.",
}
