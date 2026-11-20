package support

import "github.com/jmoiron/sqlx"

type CreatedSendEmailVerificationLinkJob struct {
	RecipientEmail   string
	VerificationLink string
}

type TestingEmailService struct {
	CreatedJobs []CreatedSendEmailVerificationLinkJob
}

func NewTestingEmailService() *TestingEmailService {
	return &TestingEmailService{}
}

func (service *TestingEmailService) CreateSendEmailVerificationLinkJob(
	_ *sqlx.Tx,
	recipientEmail string,
	emailVerificationLink string,
) error {
	service.CreatedJobs = append(service.CreatedJobs, CreatedSendEmailVerificationLinkJob{
		RecipientEmail:   recipientEmail,
		VerificationLink: emailVerificationLink,
	})
	return nil
}
