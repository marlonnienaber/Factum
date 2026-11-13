package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationVerifyEmailInvalidEmail(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: "adaexample.com",
	})
	require.NoError(t, err, "serializing verify-email request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", payload)

	support.AssertValidationErrorResponse(t, resp, "email", "email must be a valid email address.")
	assert.Empty(t, sut.EmailService.SentEmails)
}

func TestAuthenticationVerifyEmailMissingEmail(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", []byte(`{}`))
	support.AssertValidationErrorResponse(t, resp, "email", "email must be a valid email address.")
	assert.Empty(t, sut.EmailService.SentEmails)
}

func TestAuthenticationVerifyEmailInvalidJson(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", []byte(`{"email":`))
	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"json_parsing_error",
		"Request body does not represent valid JSON.",
	)
	assert.Empty(t, sut.EmailService.SentEmails)
}

func TestAuthenticationVerifyEmailEmailNotWhitelisted(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser2Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", payload)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	assert.Empty(t, sut.EmailService.SentEmails)
}

func TestAuthenticationVerifyEmailEmailAlreadyInUse(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	require.Len(t, sut.EmailService.SentEmails, 1)

	sut.Clock.Advance(time.Minute)

	payload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", payload)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	assert.Len(t, sut.EmailService.SentEmails, 1)
}

func TestAuthenticationVerifyEmailMinutelySendLimit(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	firstPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	firstResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", firstPayload)

	require.Equal(t, http.StatusNoContent, firstResp.Code)
	assert.Empty(t, firstResp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 1)

	sut.Clock.Advance(30 * time.Second)

	secondPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	secondResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", secondPayload)

	require.Equal(t, http.StatusNoContent, secondResp.Code)
	assert.Empty(t, secondResp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 1)

	sut.Clock.Advance(30 * time.Second)

	thirdPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	thirdResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", thirdPayload)

	require.Equal(t, http.StatusNoContent, thirdResp.Code)
	assert.Empty(t, thirdResp.Body.String(), "expected an empty response body")
	assert.Len(t, sut.EmailService.SentEmails, 2)
}

func TestAuthenticationVerifyEmailHourlySendLimit(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	for index := range 5 {
		firstPayload, err := json.Marshal(transport.VerifyEmailRequest{
			Email: support.TestUser1Email,
		})
		require.NoError(t, err, "serializing verify-email request failed")
		firstResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", firstPayload)

		require.Equal(t, http.StatusNoContent, firstResp.Code)
		assert.Empty(t, firstResp.Body.String(), "expected an empty response body")
		require.Len(t, sut.EmailService.SentEmails, index+1)
		sut.Clock.Advance(time.Minute)
	}

	secondPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	secondResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", secondPayload)

	require.Equal(t, http.StatusNoContent, secondResp.Code)
	assert.Empty(t, secondResp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 5)
	sut.Clock.Advance(55 * time.Minute)

	thirdPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	thirdResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", thirdPayload)

	require.Equal(t, http.StatusNoContent, thirdResp.Code)
	assert.Empty(t, thirdResp.Body.String(), "expected an empty response body")
	assert.Len(t, sut.EmailService.SentEmails, 6)
}

func TestAuthenticationVerifyEmailDailySendLimit(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	for index := range 10 {
		firstPayload, err := json.Marshal(transport.VerifyEmailRequest{
			Email: support.TestUser1Email,
		})
		require.NoError(t, err, "serializing verify-email request failed")
		firstResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", firstPayload)

		require.Equal(t, http.StatusNoContent, firstResp.Code)
		assert.Empty(t, firstResp.Body.String(), "expected an empty response body")
		require.Len(t, sut.EmailService.SentEmails, index+1)
		sut.Clock.Advance(time.Hour)
	}

	secondPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	secondResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", secondPayload)

	require.Equal(t, http.StatusNoContent, secondResp.Code)
	assert.Empty(t, secondResp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 10)

	sut.Clock.Advance(14 * time.Hour)

	thirdPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	thirdResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", thirdPayload)

	require.Equal(t, http.StatusNoContent, thirdResp.Code)
	assert.Empty(t, thirdResp.Body.String(), "expected an empty response body")
	assert.Len(t, sut.EmailService.SentEmails, 11)
}

func TestAuthenticationVerifyEmailEmailDeliveryFailure(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	sut.EmailService.SendError = errors.New("email delivery unavailable")

	payload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", payload)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	require.Empty(t, sut.EmailService.SentEmails)

}

func TestAuthenticationVerifyEmailSuccess(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", payload)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 1)
	sentEmail := sut.EmailService.SentEmails[0]
	assert.Equal(t, support.TestUser1Email, sentEmail.RecipientEmail)

	link, err := url.Parse(sentEmail.VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	assert.Equal(t, "http", link.Scheme)
	assert.Equal(t, "localhost:5173", link.Host)
	assert.Equal(t, "/sign-up", link.Path)

	token := link.Query().Get("token")
	require.NotEmpty(t, token, "verification link did not contain a token")
}

func TestAuthenticationVerifyEmailNormalizesEmail(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	payload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: strings.ToUpper(support.TestUser1Email),
	})
	require.NoError(t, err, "serializing verify-email request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", payload)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 1)
	assert.Equal(t, support.TestUser1Email, sut.EmailService.SentEmails[0].RecipientEmail)

}

func TestAuthenticationVerifyEmailResendInvalidatesPreviousToken(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App

	firstPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	firstResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", firstPayload)

	require.Equal(t, http.StatusNoContent, firstResp.Code)
	assert.Empty(t, firstResp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 1)

	previousTokenLink, err := url.Parse(sut.EmailService.SentEmails[0].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	previousToken := previousTokenLink.Query().Get("token")
	require.NotEmpty(t, previousToken, "verification link did not contain a token")

	sut.Clock.Advance(time.Minute)

	secondPayload, err := json.Marshal(transport.VerifyEmailRequest{
		Email: support.TestUser1Email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	secondResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/verify-email", secondPayload)

	require.Equal(t, http.StatusNoContent, secondResp.Code)
	assert.Empty(t, secondResp.Body.String(), "expected an empty response body")
	require.Len(t, sut.EmailService.SentEmails, 2)

	newTokenLink, err := url.Parse(sut.EmailService.SentEmails[1].VerificationLink)
	require.NoError(t, err, "parsing verification link failed")
	newToken := newTokenLink.Query().Get("token")
	require.NotEmpty(t, newToken, "verification link did not contain a token")
	require.NotEqual(t, previousToken, newToken)

	signUpPayload, err := json.Marshal(transport.SignUpRequest{
		Token:    previousToken,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", signUpPayload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"sign_up_not_available",
		"Sign-up is not available. Please restart the sign-up process.",
	)

	newSignUpPayload, err := json.Marshal(transport.SignUpRequest{
		Token:    newToken,
		Password: support.TestUser1Password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	signUpResp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", newSignUpPayload)

	require.Equal(t, http.StatusCreated, signUpResp.Code)
}
