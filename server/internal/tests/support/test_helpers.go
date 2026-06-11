package support

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
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

func MakeSignUpRequest(t *testing.T, app *gin.Engine, body transport.AuthenticationRequest) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	return MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)
}

func MakeJsonRequest(t *testing.T, app *gin.Engine, method string, path string, payload []byte) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

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
