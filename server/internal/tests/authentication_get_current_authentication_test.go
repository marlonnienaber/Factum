package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationGetCurrentAuthenticationSuccess(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)

	signUpResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/current-authentication",
		nil,
		signUpResp.Result().Cookies()...,
	)

	require.Equal(t, http.StatusOK, resp.Code)

	var respBody transport.AuthenticationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")
	assert.Equal(t, int64(1), respBody.UserId, "unexpected userId")
}

func TestAuthenticationGetCurrentAuthenticationNoUserForSession(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)

	signUpResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	_, err := support.DB.Exec("DELETE FROM users WHERE id = $1", int64(1))
	require.NoError(t, err, "failed to delete user while preparing database state")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/current-authentication",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"no_user_for_session",
		"The user account for this session was deleted. Please sign-in with a different account.",
	)
}

func TestAuthenticationGetCurrentAuthenticationUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/current-authentication", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}
