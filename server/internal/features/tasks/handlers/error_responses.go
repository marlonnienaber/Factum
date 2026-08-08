package handlers

import "github.com/marlonnienaber/Factum/server/internal/shared"

var invalidQueryParameterCombinationErrorResponse = shared.ErrorResponse{
	ErrorCode: "invalid_query_parameter_combination",
	Message:   "URL contains an invalid query parameter combination.",
}
