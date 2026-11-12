package tests

import "testing"

func TestAuthenticationVerifyEmailInvalidEmail(t *testing.T) {}

func TestAuthenticationVerifyEmailMissingEmail(t *testing.T) {}

func TestAuthenticationVerifyEmailInvalidJson(t *testing.T) {}

func TestAuthenticationVerifyEmailEmailNotWhitelisted(t *testing.T) {}

func TestAuthenticationVerifyEmailEmailAlreadyInUse(t *testing.T) {}

func TestAuthenticationVerifyEmailResendTooSoon(t *testing.T) {}

func TestAuthenticationVerifyEmailHourlySendLimitReached(t *testing.T) {}

func TestAuthenticationVerifyEmailDailySendLimitReached(t *testing.T) {}

func TestAuthenticationVerifyEmailEmailDeliveryFailure(t *testing.T) {}

func TestAuthenticationVerifyEmailSuccess(t *testing.T) {}

func TestAuthenticationVerifyEmailNormalizesEmail(t *testing.T) {}

func TestAuthenticationVerifyEmailResendSuccess(t *testing.T) {}

func TestAuthenticationVerifyEmailResendInvalidatesPreviousToken(t *testing.T) {}
