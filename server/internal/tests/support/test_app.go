package support

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/application"
	"github.com/marlonnienaber/Factum/server/internal/injections"
)

type SystemUnderTest struct {
	App          *gin.Engine
	EmailService *TestingEmailService
	Clock        *TestingClock
}

func NewSystemUnderTest(env injections.Environment) SystemUnderTest {
	emailService := NewTestingEmailService()
	clock := NewTestingClock(time.Now())
	app := application.NewApplication(
		env,
		DB,
		application.WithEmailService(emailService),
		application.WithClock(clock),
	)

	return SystemUnderTest{
		App:          app,
		EmailService: emailService,
		Clock:        clock,
	}
}
