package services

import "log/slog"

type EmailService interface {
	SendEmailVerificationLink(recipientEmail string, emailVerificationLink string) error
}

type DevelopmentEmailService struct{}

func NewDevelopmentEmailService() *DevelopmentEmailService {
	return &DevelopmentEmailService{}
}

func (service *DevelopmentEmailService) SendEmailVerificationLink(
	recipientEmail string,
	emailVerificationLink string,
) error {
	slog.Info(
		"Sending email verification link with development email service.",
		"recipientEmail", recipientEmail,
		"emailVerificationLink", emailVerificationLink,
	)
	return nil
}
