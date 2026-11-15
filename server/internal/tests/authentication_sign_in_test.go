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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	support.MakeVerifyEmailAndSignUpRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	body := transport.SignInRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	body := transport.SignInRequest{
		Email:    "benexample.com",
		Password: support.TestUser1Password,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertValidationErrorResponse(t, resp, "email", "email must be a valid email address.")
}

func TestAuthenticationSignInPasswordTooShort(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	body := transport.SignInRequest{
		Email:    support.TestUser1Email,
		Password: strings.Repeat("a", 9),
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at least 10 characters long.")
}

func TestAuthenticationSignInPasswordTooLong(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	body := transport.SignInRequest{
		Email:    support.TestUser1Email,
		Password: strings.Repeat("a", 129),
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-in", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at most 128 characters long.")
}

func TestAuthenticationSignInInvalidCredentialsOnNonExistentEmail(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	body := transport.SignInRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	support.MakeVerifyEmailAndSignUpRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	body := transport.SignInRequest{
		Email:    support.TestUser1Email,
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
