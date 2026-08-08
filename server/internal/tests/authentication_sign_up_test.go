package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationSignUpSuccess(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	resp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    strings.ToUpper(support.TestUserEmail),
		Password: support.TestUserPassword,
	})

	require.Equal(t, http.StatusCreated, resp.Code)

	var respBody transport.AuthenticationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")
	assert.Equal(t, int64(1), respBody.UserId, "unexpected userId")
	assert.NotEmpty(t, resp.Header().Get("Set-Cookie"), "expected sign-up to create a session cookie")

	var persistedEmail string
	require.NoError(t, support.DB.Get(&persistedEmail, "SELECT email FROM users WHERE id = $1", respBody.UserId))
	assert.Equal(t, support.TestUserEmail, persistedEmail, "expected email to be normalized before persistence")
}

func TestAuthenticationSignUpInvalidEmail(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	resp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    "adaexample.com",
		Password: support.TestUserPassword,
	})

	support.AssertValidationErrorResponse(t, resp, "email", "email must be a valid email address.")
}

func TestAuthenticationSignUpPasswordTooShort(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	resp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: strings.Repeat("a", 9),
	})

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at least 10 characters long.")
}

func TestAuthenticationSignUpPasswordTooLong(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	resp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: strings.Repeat("a", 129),
	})

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at most 128 characters long.")
}

func TestAuthenticationSignUpInvalidJson(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", []byte(`{"email":`))

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"json_parsing_error",
		"Request body does not represent valid JSON.",
	)
}

func TestAuthenticationSignUpEmailNotAllowed(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, "allowed@example.com")

	resp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    "blocked@example.com",
		Password: support.TestUserPassword,
	})

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusForbidden,
		"email_not_allowed",
		"You are not allowed to sign up with this email.",
	)
}

func TestAuthenticationSignUpEmailAlreadyInUse(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	firstResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, firstResp.Code)

	secondResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})

	support.AssertErrorResponse(
		t,
		secondResp,
		http.StatusConflict,
		"email_already_in_use",
		"Email is already in use.",
	)
}
