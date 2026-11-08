package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/jmoiron/sqlx"
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

const (
	emailVerificationTokenByteCount       = 32
	emailVerificationTokenLifetime        = 15 * time.Minute
	minimumEmailVerificationSendInterval  = time.Minute
	emailVerificationHourlyWindowDuration = time.Hour
	emailVerificationDailyWindowDuration  = 24 * time.Hour
	maximumEmailVerificationHourlySends   = 5
	maximumEmailVerificationDailySends    = 10
)

type AuthenticationService struct {
	db                                *sqlx.DB
	userStore                         *stores.UserStore
	emailVerificationTokenRecordStore *stores.EmailVerificationTokenRecordStore
	emailService                      EmailService
	cfg                               configuration.Configuration
}

func NewAuthenticationService(
	db *sqlx.DB,
	userStore *stores.UserStore,
	emailVerificationTokenRecordStore *stores.EmailVerificationTokenRecordStore,
	emailService EmailService,
	config configuration.Configuration,
) *AuthenticationService {
	return &AuthenticationService{
		db:                                db,
		userStore:                         userStore,
		emailVerificationTokenRecordStore: emailVerificationTokenRecordStore,
		emailService:                      emailService,
		cfg:                               config,
	}
}

func (service *AuthenticationService) VerifyEmail(
	request transport.VerifyEmailRequest,
) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("verifying email failed")
	email := normalizeEmail(request.Email)

	if service.isEmailNotWhitelisted(email) {
		return nil
	}

	plaintextToken, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (string, error) {
		userExists, err := service.userStore.DoesUserExistForEmail(transaction, email)
		if err != nil {
			return "", err
		}
		if userExists {
			return "", nil
		}

		record, err := service.emailVerificationTokenRecordStore.GetEmailVerificationTokenRecordForEmail(
			transaction,
			email,
		)
		if err != nil && !errors.Is(err, stores.ErrNoEmailVerificationTokenRecordForEmail) {
			return "", err
		}

		now := time.Now().UTC()
		if record != nil && !isEmailVerificationSendAllowed(record, now) {
			return "", nil
		}

		plaintextToken, err := generateEmailVerificationToken()
		if err != nil {
			return "", err
		}
		tokenHash := sha256.Sum256([]byte(plaintextToken))

		record = prepareEmailVerificationTokenRecord(record, email, tokenHash[:], now)
		err = service.emailVerificationTokenRecordStore.UpsertEmailVerificationTokenRecord(transaction, record)
		if err != nil {
			return "", err
		}

		return plaintextToken, nil
	})
	if err != nil {
		return wrapUnexpectedError(err)
	}
	if plaintextToken == "" {
		return nil
	}

	emailVerificationLink := service.cfg.AppUrl + "/sign-up?token=" + plaintextToken
	if err := service.emailService.SendEmailVerificationLink(email, emailVerificationLink); err != nil {
		slog.Error("sending email verification link failed.", "error", err)
	}

	return nil
}

func (service *AuthenticationService) SignUp(request transport.AuthenticationRequest, session sessions.Session) (*transport.AuthenticationResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("sign up failed")
	email := normalizeEmail(request.Email)

	if service.isEmailNotWhitelisted(email) {
		return nil, ErrEmailNotAllowed
	}

	user, err := service.createUser(email, request.Password)
	if err != nil {
		if errors.Is(err, stores.ErrEmailAlreadyInUse) {
			return nil, err
		}
		return nil, wrapUnexpectedError(err)
	}

	err = service.createSession(user.Id, session)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	authenticationResponse := user.ToAuthenticationResponse()
	return &authenticationResponse, nil
}

func (service *AuthenticationService) SignIn(request transport.AuthenticationRequest, session sessions.Session) (*transport.AuthenticationResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("sign in failed")
	email := normalizeEmail(request.Email)

	var user *domain.User
	var fakeHash = "$argon2id$v=19$m=65536,t=1,p=8$Sa6aciSgI6dRYahIkrqpSA$i5itaM8W0RhBvuPNrYhbXsLe3JQViMZ8FWf8gQKz52k"
	var hashToTest = fakeHash

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

	err = service.createSession(user.Id, session)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	authenticationResponse := user.ToAuthenticationResponse()
	return &authenticationResponse, nil
}

func (service *AuthenticationService) DeleteSession(session sessions.Session) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("session deletion failed")

	session.Clear()
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: -1,
	})
	if err := session.Save(); err != nil {
		return wrapUnexpectedError(err)
	}

	return nil
}

func (service *AuthenticationService) createSession(userId int64, session sessions.Session) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("session creation failed")

	session.Set("userId", userId)
	if err := session.Save(); err != nil {
		return wrapUnexpectedError(err)
	}

	return nil
}

func (service *AuthenticationService) isEmailNotWhitelisted(email string) bool {
	for _, allowedEmail := range service.cfg.SignUpEmailWhitelist {
		if allowedEmail == email {
			return false
		}
	}
	return true
}

func (service *AuthenticationService) createUser(email string, password string) (*domain.User, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("user creation failed")

	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
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

func normalizeEmail(email string) string {
	return strings.ToLower(email)
}

func generateEmailVerificationToken() (string, error) {
	plaintextTokenBytes := make([]byte, emailVerificationTokenByteCount)
	if _, err := rand.Read(plaintextTokenBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(plaintextTokenBytes), nil
}

func isEmailVerificationSendAllowed(record *domain.EmailVerificationTokenRecord, now time.Time) bool {
	if now.Before(record.LastSentAt.Add(minimumEmailVerificationSendInterval)) {
		return false
	}

	hourlyWindowActive := now.Before(record.HourlyWindowStartedAt.Add(emailVerificationHourlyWindowDuration))
	if hourlyWindowActive && record.HourlySendCount >= maximumEmailVerificationHourlySends {
		return false
	}

	dailyWindowActive := now.Before(record.DailyWindowStartedAt.Add(emailVerificationDailyWindowDuration))
	if dailyWindowActive && record.DailySendCount >= maximumEmailVerificationDailySends {
		return false
	}

	return true
}

func prepareEmailVerificationTokenRecord(
	record *domain.EmailVerificationTokenRecord,
	email string,
	tokenHash []byte,
	now time.Time,
) *domain.EmailVerificationTokenRecord {
	if record == nil {
		return &domain.EmailVerificationTokenRecord{
			Email:                 email,
			TokenHash:             tokenHash,
			ExpiresAt:             now.Add(emailVerificationTokenLifetime),
			LastSentAt:            now,
			HourlySendCount:       1,
			HourlyWindowStartedAt: now,
			DailySendCount:        1,
			DailyWindowStartedAt:  now,
		}
	}

	record.TokenHash = tokenHash
	record.ExpiresAt = now.Add(emailVerificationTokenLifetime)
	record.LastSentAt = now

	if now.Before(record.HourlyWindowStartedAt.Add(emailVerificationHourlyWindowDuration)) {
		record.HourlySendCount++
	} else {
		record.HourlySendCount = 1
		record.HourlyWindowStartedAt = now
	}

	if now.Before(record.DailyWindowStartedAt.Add(emailVerificationDailyWindowDuration)) {
		record.DailySendCount++
	} else {
		record.DailySendCount = 1
		record.DailyWindowStartedAt = now
	}

	return record
}
