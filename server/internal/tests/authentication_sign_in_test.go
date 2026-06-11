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

func TestAuthenticationSignInSuccess(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	body := transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	require.Equal(t, http.StatusOK, resp.Code)

	var respBody transport.AuthenticationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")
	assert.Equal(t, int64(1), respBody.UserId, "unexpected userId")
	assert.NotEmpty(t, resp.Header().Get("Set-Cookie"), "expected sign-in to create a session cookie")
}

func TestAuthenticationSignInInvalidEmail(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	body := transport.AuthenticationRequest{
		Email:    "benexample.com",
		Password: support.TestUserPassword,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertValidationErrorResponse(t, resp, "email", "email must be a valid email address.")
}

func TestAuthenticationSignInPasswordTooShort(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	body := transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: strings.Repeat("a", 9),
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at least 10 characters long.")
}

func TestAuthenticationSignInPasswordTooLong(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	body := transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: strings.Repeat("a", 129),
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at most 128 characters long.")
}

func TestAuthenticationSignInInvalidJson(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", []byte(`{"email":`))

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"json_parsing_error",
		"Request body does not represent valid JSON.",
	)
}

func TestAuthenticationSignInInvalidCredentialsOnNonExistentEmail(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	body := transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"invalid_credentials",
		"Email or password incorrect.",
	)
}

func TestAuthenticationSignInInvalidCredentialsOnWrongPassword(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	body := transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: "wrongPassword",
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"invalid_credentials",
		"Email or password incorrect.",
	)
}
