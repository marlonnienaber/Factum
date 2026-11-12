package support

type SentEmailInfo struct {
	RecipientEmail   string
	VerificationLink string
}

type TestingEmailService struct {
	SentEmails []SentEmailInfo
	SendError  error
}

func NewTestingEmailService() *TestingEmailService {
	return &TestingEmailService{}
}

func (service *TestingEmailService) SendEmailVerificationLink(
	recipientEmail string,
	emailVerificationLink string,
) error {
	if service.SendError != nil {
		return service.SendError
	}

	service.SentEmails = append(service.SentEmails, SentEmailInfo{
		RecipientEmail:   recipientEmail,
		VerificationLink: emailVerificationLink,
	})
	return nil
}
