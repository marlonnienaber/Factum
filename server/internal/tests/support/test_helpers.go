package support

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func PrepareTestEnvironment(t *testing.T) *gin.Engine {
	t.Helper()
	return prepareTestEnvironmentWithConfig(t, defaultTestConfiguration)
}

func PrepareTestEnvironmentWithSignUpEmailWhitelist(t *testing.T, emails ...string) *gin.Engine {
	t.Helper()
	cfg := defaultTestConfiguration
	cfg.SignUpEmailWhitelist = emails
	return prepareTestEnvironmentWithConfig(t, cfg)
}

func prepareTestEnvironmentWithConfig(t *testing.T, cfg configuration.Configuration) *gin.Engine {
	t.Helper()
	databaseReset(t)
	return TestApp(cfg)
}

var defaultTestConfiguration = configuration.Configuration{
	Port:                ":3001",
	SessionCookieSecure: false,
	SessionSecretKey:    "test-session-secret-key",
}

const (
	TestUserEmail    = "ada@example.com"
	TestUserPassword = "passwordPassword"
)

func MakeSignUpRequest(t *testing.T, app *gin.Engine, body authenticationTransport.AuthenticationRequest) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	return MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)
}

func MakeCreateCategoryRequest(
	t *testing.T,
	app *gin.Engine,
	title string,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	path := "/api/categories?" + url.Values{"title": {title}}.Encode()
	return MakeJsonRequest(t, app, http.MethodPost, path, nil, cookies...)
}

func MakeCreateTaskRequest(
	t *testing.T,
	app *gin.Engine,
	title string,
	description *string,
	date *shared.Date,
	categoryId int64,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:       title,
		Description: description,
		Date:        date,
		CategoryId:  categoryId,
	})
	require.NoError(t, err, "serializing request failed")

	return MakeJsonRequest(t, app, http.MethodPost, "/api/tasks", payload, cookies...)
}

func MakeJsonRequest(
	t *testing.T,
	app *gin.Engine,
	method string,
	path string,
	payload []byte,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	return resp
}

func AssertValidationErrorResponse(
	t *testing.T,
	resp *httptest.ResponseRecorder,
	expectedField string,
	expectedViolationDescription string,
) {
	t.Helper()

	require.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody shared.DetailedErrorResponse[validation.ValidationErrorDetails]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	require.Equal(t, "validation_error", respBody.ErrorCode, "unexpected errorCode")
	require.Equal(t, "Request body did not match the expected form.", respBody.Message, "unexpected message")
	require.Len(t, respBody.Details.Violations, 1, "expected exactly one violation")

	violation := respBody.Details.Violations[0]
	assert.Equal(t, expectedField, violation.Field, "unexpected field violated")
	assert.Equal(t, expectedViolationDescription, violation.Description, "unexpected violation description")
}

func AssertErrorResponse(
	t *testing.T,
	resp *httptest.ResponseRecorder,
	expectedStatusCode int,
	expectedErrorCode string,
	expectedMessage string,
) {
	t.Helper()

	require.Equal(t, expectedStatusCode, resp.Code)

	var respBody shared.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, expectedErrorCode, respBody.ErrorCode, "unexpected error code")
	assert.Equal(t, expectedMessage, respBody.Message, "unexpected message")
}
