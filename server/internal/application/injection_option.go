package application

import (
	authServices "github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

type injectableAppComponents struct {
	emailService authServices.EmailService
	clock        shared.Clock
}

type InjectionOption interface {
	inject(*injectableAppComponents)
}

type emailServiceInjectionOption struct {
	emailService authServices.EmailService
}

func (option emailServiceInjectionOption) inject(components *injectableAppComponents) {
	components.emailService = option.emailService
}

func WithEmailService(emailService authServices.EmailService) InjectionOption {
	if emailService == nil {
		panic("email service must not be nil")
	}

	return emailServiceInjectionOption{emailService: emailService}
}

type clockInjectionOption struct {
	clock shared.Clock
}

func (option clockInjectionOption) inject(components *injectableAppComponents) {
	components.clock = option.clock
}

func WithClock(clock shared.Clock) InjectionOption {
	if clock == nil {
		panic("clock must not be nil")
	}

	return clockInjectionOption{clock: clock}
}
