package services

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type EmailService interface {
	CreateSendEmailVerificationLinkJob(transaction *sqlx.Tx, recipientEmail string, emailVerificationLink string) error
}

type DevelopmentEmailService struct{}

func NewDevelopmentEmailService() *DevelopmentEmailService {
	return &DevelopmentEmailService{}
}

func (service *DevelopmentEmailService) CreateSendEmailVerificationLinkJob(
	_ *sqlx.Tx,
	recipientEmail string,
	emailVerificationLink string,
) error {
	slog.Info(
		"Creating email verification link job with development email service.",
		"recipientEmail", recipientEmail,
		"emailVerificationLink", emailVerificationLink,
	)
	return nil
}
