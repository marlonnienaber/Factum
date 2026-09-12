package shared

type ErrorResponse struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}

type DetailedErrorResponse[T any] struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Details   T      `json:"details,omitempty"`
}

var JsonParsingErrorResponse = ErrorResponse{
	ErrorCode: "json_parsing_error",
	Message:   "Request body does not represent valid JSON.",
}

var InvalidQueryParameterErrorResponse = ErrorResponse{
	ErrorCode: "invalid_query_parameter",
	Message:   "URL contains an invalid query parameter",
}

var MissingQueryParameterErrorResponse = ErrorResponse{
	ErrorCode: "missing_query_parameter",
	Message:   "URL is missing a required query parameter.",
}

var UnexpectedErrorResponse = ErrorResponse{
	ErrorCode: "unexpected_error",
	Message:   "An unexpected error occurred.",
}
