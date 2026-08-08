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

var UnexpectedErrorResponse = ErrorResponse{
	ErrorCode: "unexpected_error",
	Message:   "An unexpected error occurred.",
}
