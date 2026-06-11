package tests

import (
	"net/http"
	"testing"

	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationSignOutSuccess(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)

	signUpResp := support.MakeSignUpRequest(t, app, transport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	signOutResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/sign-out",
		nil,
		signUpResp.Result().Cookies()...,
	)

	require.Equal(t, http.StatusNoContent, signOutResp.Code)
	require.NotEmpty(t, signOutResp.Header().Get("Set-Cookie"), "expected sign-out to clear the session cookie")
}

func TestAuthenticationSignOutUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-out", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}
