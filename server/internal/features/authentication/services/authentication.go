package services

import (
	"errors"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var (
	ErrEmailNotAllowed    = errors.New("the provided email is not whitelisted.")
	ErrInvalidCredentials = errors.New("username or password invalid")
)

type AuthenticationService struct {
	userStore *stores.UserStore
	cfg       configuration.Configuration
}

func NewAuthenticationService(userStore *stores.UserStore, config configuration.Configuration) *AuthenticationService {
	return &AuthenticationService{
		userStore: userStore,
		cfg:       config,
	}
}

func (service *AuthenticationService) SignUp(request transport.AuthenticationRequest, session sessions.Session) (*transport.AuthenticationResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("sign up failed")

	if service.isEmailNotWhitelisted(request.Email) {
		return nil, ErrEmailNotAllowed
	}

	user, err := service.createUser(request)
	if err != nil {
		if errors.Is(err, stores.ErrEmailAlreadyInUse) {
			return nil, err
		}
		return nil, wrapUnexpectedError(err)
	}

	session.Set("userId", user.Id)
	if err := session.Save(); err != nil {
		return nil, wrapUnexpectedError(err)
	}

	authenticationResponse := user.ToAuthenticationResponse()
	return &authenticationResponse, nil
}

func (service *AuthenticationService) SignIn(request transport.AuthenticationRequest, session sessions.Session) (*transport.AuthenticationResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("sign in failed")

	var user *domain.User
	var fakeHash = "$argon2id$v=19$m=65536,t=1,p=8$Sa6aciSgI6dRYahIkrqpSA$i5itaM8W0RhBvuPNrYhbXsLe3JQViMZ8FWf8gQKz52k"
	var hashToTest = fakeHash
	email := strings.ToLower(request.Email)

	user, err := service.userStore.GetUserByEmail(email)
	if err != nil {
		if !errors.Is(err, stores.ErrNoUserForGivenEmail) {
			return nil, wrapUnexpectedError(err)
		}
	} else {
		hashToTest = user.PasswordHash
	}

	match, err := argon2id.ComparePasswordAndHash(request.Password, hashToTest)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	if user == nil || !match {
		return nil, ErrInvalidCredentials
	}

	session.Set("userId", user.Id)
	if err := session.Save(); err != nil {
		return nil, wrapUnexpectedError(err)
	}

	authenticationResponse := user.ToAuthenticationResponse()
	return &authenticationResponse, nil
}

func (service *AuthenticationService) isEmailNotWhitelisted(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	for _, allowedEmail := range service.cfg.SignUpEmailWhitelist {
		if allowedEmail == email {
			return false
		}
	}
	return true
}

func (service *AuthenticationService) createUser(request transport.AuthenticationRequest) (*domain.User, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("user creation failed")

	email := strings.ToLower(request.Email)

	passwordHash, err := argon2id.CreateHash(request.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	user := domain.User{
		Email:        email,
		PasswordHash: passwordHash,
	}
	err = service.userStore.InsertUser(&user)
	if err != nil {
		if errors.Is(err, stores.ErrEmailAlreadyInUse) {
			return nil, err
		}
		return nil, wrapUnexpectedError(err)
	}

	return &user, nil
}
