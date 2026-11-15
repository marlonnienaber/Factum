package tests

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/marlonnienaber/Factum/server/internal/application"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationSignUpSuccess(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	require.Equal(t, http.StatusCreated, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	assert.Empty(t, resp.Header().Get("Set-Cookie"), "sign-up must not create a session")
}

func TestAuthenticationSignUpSuccessWithMinimumPasswordLength(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: strings.Repeat("a", 10),
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	require.Equal(t, http.StatusCreated, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	assert.Empty(t, resp.Header().Get("Set-Cookie"), "sign-up must not create a session")
}

func TestAuthenticationSignUpSuccessWithMaximumPasswordLength(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: strings.Repeat("a", 128),
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	require.Equal(t, http.StatusCreated, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	assert.Empty(t, resp.Header().Get("Set-Cookie"), "sign-up must not create a session")
}

func TestAuthenticationSignUpSuccessJustBeforeTokenExpiry(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	sut.Clock.Set(time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC))
	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")
	sut.Clock.Advance(15*time.Minute - time.Microsecond)

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	require.Equal(t, http.StatusCreated, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	assert.Empty(t, resp.Header().Get("Set-Cookie"), "sign-up must not create a session")
}

func TestAuthenticationSignUpMissingToken(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(map[string]string{"password": support.TestUser1Password})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "token", "token must be a valid email verification token.")
}

func TestAuthenticationSignUpTokenTooShort(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    strings.Repeat("a", 63),
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "token", "token must be a valid email verification token.")
}

func TestAuthenticationSignUpTokenTooLong(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    strings.Repeat("a", 65),
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "token", "token must be a valid email verification token.")
}

func TestAuthenticationSignUpTokenInvalidCharacters(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    strings.Repeat("a", 63) + "g",
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "token", "token must be a valid email verification token.")
}

func TestAuthenticationSignUpMissingPassword(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	payload, err := json.Marshal(map[string]string{"token": token})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at least 10 characters long.")
}

func TestAuthenticationSignUpPasswordTooShort(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: strings.Repeat("a", 9),
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at least 10 characters long.")
}

func TestAuthenticationSignUpPasswordTooLong(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: strings.Repeat("a", 129),
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertValidationErrorResponse(t, resp, "password", "password must be at most 128 characters long.")
}

func TestAuthenticationSignUpUnknownToken(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    strings.Repeat("a", 64),
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"sign_up_not_available",
		"Sign-up is not available. Please restart the sign-up process.",
	)
}

func TestAuthenticationSignUpTokenAtExpiry(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	sut.Clock.Set(time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC))
	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")
	sut.Clock.Advance(15 * time.Minute)

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"sign_up_not_available",
		"Sign-up is not available. Please restart the sign-up process.",
	)
}

func TestAuthenticationSignUpConsumedToken(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	firstPayload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing first sign-up request failed")
	firstResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", firstPayload)
	require.Equal(t, http.StatusCreated, firstResp.Code)

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"sign_up_not_available",
		"Sign-up is not available. Please restart the sign-up process.",
	)
}

func TestAuthenticationSignUpEmailNoLongerWhitelisted(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	app := application.NewApplication(
		configuration.Configuration{
			SignUpEmailWhitelist: []string{support.TestUser2Email},
			SessionSecretKey:     "test-session-secret-key",
		},
		support.DB,
		application.WithEmailService(sut.EmailService),
		application.WithClock(sut.Clock),
	)

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"sign_up_not_available",
		"Sign-up is not available. Please restart the sign-up process.",
	)
}

func TestAuthenticationSignUpEmailAlreadyInUse(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	verifyEmailResp := support.MakeVerifyEmailRequest(t, sut, support.TestUser1Email)
	require.Equal(t, http.StatusNoContent, verifyEmailResp.Code)
	require.Len(t, sut.EmailService.SentEmails, 1)
	verificationLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	token := verificationLink.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")

	_, err = support.DB.Exec(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		support.TestUser1Email,
		"unused-password-hash",
	)
	require.NoError(t, err, "creating existing account failed")

	payload, err := json.Marshal(transport.SignUpRequest{
		Token:    token,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"sign_up_not_available",
		"Sign-up is not available. Please restart the sign-up process.",
	)
}
