package tests

import (
	"net/http"
	"testing"

	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationSignOutSuccess(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	signInResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	signOutResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/sign-out",
		nil,
		signInResp.Result().Cookies()...,
	)

	require.Equal(t, http.StatusNoContent, signOutResp.Code)
	require.NotEmpty(t, signOutResp.Header().Get("Set-Cookie"), "expected sign-out to clear the session cookie")
}

func TestAuthenticationSignOutUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-out", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}
