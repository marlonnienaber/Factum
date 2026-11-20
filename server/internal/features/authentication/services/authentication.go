package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/injections"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var (
	ErrSignUpNotAvailable = errors.New("sign-up is not available")
	ErrInvalidCredentials = errors.New("username or password invalid")
)

const (
	emailVerificationTokenByteCount         = 32
	emailVerificationTokenLifetime          = 15 * time.Minute
	emailVerificationMinutelyWindowDuration = time.Minute
	emailVerificationHourlyWindowDuration   = time.Hour
	emailVerificationDailyWindowDuration    = 24 * time.Hour
	maximumEmailVerificationHourlySends     = 5
	maximumEmailVerificationDailySends      = 10
)

type AuthenticationService struct {
	db                                *sqlx.DB
	userStore                         *stores.UserStore
	emailVerificationTokenRecordStore *stores.EmailVerificationTokenRecordStore
	emailService                      EmailService
	clock                             shared.Clock
	env                               injections.Environment
}

func NewAuthenticationService(
	db *sqlx.DB,
	userStore *stores.UserStore,
	emailVerificationTokenRecordStore *stores.EmailVerificationTokenRecordStore,
	emailService EmailService,
	clock shared.Clock,
	env injections.Environment,
) *AuthenticationService {
	return &AuthenticationService{
		db:                                db,
		userStore:                         userStore,
		emailVerificationTokenRecordStore: emailVerificationTokenRecordStore,
		emailService:                      emailService,
		clock:                             clock,
		env:                               env,
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

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		userExists, err := service.userStore.DoesUserExistForEmail(transaction, email)
		if err != nil {
			return err
		}
		if userExists {
			return nil
		}

		record, err := service.emailVerificationTokenRecordStore.GetRecordForEmail(transaction, email)
		if err != nil {
			return err
		}

		now := service.clock.Now()
		if record != nil && !isEmailVerificationSendAllowed(record, now) {
			return nil
		}

		plaintextToken, err := generateEmailVerificationToken()
		if err != nil {
			return err
		}
		tokenHash := sha256.Sum256([]byte(plaintextToken))

		record = prepareEmailVerificationTokenRecord(record, email, tokenHash[:], now)
		err = service.emailVerificationTokenRecordStore.UpsertRecord(transaction, record)
		if err != nil {
			return err
		}

		emailVerificationLink := service.env.AppUrl + "/sign-up?token=" + plaintextToken
		return service.emailService.CreateSendEmailVerificationLinkJob(transaction, email, emailVerificationLink)
	})
	if err != nil {
		return wrapUnexpectedError(err)
	}

	return nil
}

func (service *AuthenticationService) SignUp(request transport.SignUpRequest) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("sign up failed")
	tokenHash := sha256.Sum256([]byte(request.Token))

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		record, err := service.emailVerificationTokenRecordStore.GetRecordForTokenHash(transaction, tokenHash[:])
		if err != nil {
			if errors.Is(err, stores.ErrNoRecordForEmailVerificationToken) {
				return ErrSignUpNotAvailable
			}
			return err
		}
		if !record.ExpiresAt.After(service.clock.Now()) {
			return ErrSignUpNotAvailable
		}

		if service.isEmailNotWhitelisted(record.Email) {
			return ErrSignUpNotAvailable
		}

		passwordHash, err := argon2id.CreateHash(request.Password, argon2id.DefaultParams)
		if err != nil {
			return err
		}

		err = service.emailVerificationTokenRecordStore.DeleteRecord(transaction, record.Id)
		if err != nil {
			return err
		}

		user := domain.User{
			Email:        record.Email,
			PasswordHash: passwordHash,
		}
		if err := service.userStore.InsertUser(transaction, &user); err != nil {
			if errors.Is(err, stores.ErrEmailAlreadyInUse) {
				return ErrSignUpNotAvailable
			}
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrSignUpNotAvailable) {
			return err
		}
		return wrapUnexpectedError(err)
	}

	return nil
}

func (service *AuthenticationService) SignIn(request transport.SignInRequest, session sessions.Session) (*transport.AuthenticationResponse, error) {
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
	for _, allowedEmail := range service.env.SignUpEmailWhitelist {
		if allowedEmail == email {
			return false
		}
	}
	return true
}

func normalizeEmail(email string) string {
	return strings.ToLower(email)
}

func generateEmailVerificationToken() (string, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("generating email verification token failed")

	plaintextTokenBytes := make([]byte, emailVerificationTokenByteCount)
	if _, err := rand.Read(plaintextTokenBytes); err != nil {
		return "", wrapUnexpectedError(err)
	}

	return hex.EncodeToString(plaintextTokenBytes), nil
}

func isEmailVerificationSendAllowed(record *domain.EmailVerificationTokenRecord, now time.Time) bool {
	minutelyWindowActive := now.Before(record.MinutelyWindowStartedAt.Add(emailVerificationMinutelyWindowDuration))
	if minutelyWindowActive {
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
			Email:                   email,
			TokenHash:               tokenHash,
			ExpiresAt:               now.Add(emailVerificationTokenLifetime),
			MinutelyWindowStartedAt: now,
			HourlySendCount:         1,
			HourlyWindowStartedAt:   now,
			DailySendCount:          1,
			DailyWindowStartedAt:    now,
		}
	}

	record.TokenHash = tokenHash
	record.ExpiresAt = now.Add(emailVerificationTokenLifetime)
	record.MinutelyWindowStartedAt = now

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
