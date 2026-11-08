package stores

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var ErrNoEmailVerificationTokenRecordForEmail = errors.New("no email verification token record exists for the given email")

type EmailVerificationTokenRecordStore struct {
	db *sqlx.DB
}

func NewEmailVerificationTokenRecordStore(db *sqlx.DB) *EmailVerificationTokenRecordStore {
	return &EmailVerificationTokenRecordStore{
		db: db,
	}
}

func (store *EmailVerificationTokenRecordStore) GetEmailVerificationTokenRecordForEmail(
	transaction *sqlx.Tx,
	email string,
) (*domain.EmailVerificationTokenRecord, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving email verification token record failed")

	var record domain.EmailVerificationTokenRecord
	err := transaction.Get(&record, `
		SELECT id,
		       email,
		       token_hash,
		       expires_at,
		       last_sent_at,
		       hourly_send_count,
		       hourly_window_started_at,
		       daily_send_count,
		       daily_window_started_at
		FROM email_verification_token_records
		WHERE email = $1
	`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoEmailVerificationTokenRecordForEmail
		}
		return nil, wrapUnexpectedError(err)
	}

	return &record, nil
}

func (store *EmailVerificationTokenRecordStore) UpsertEmailVerificationTokenRecord(
	transaction *sqlx.Tx,
	record *domain.EmailVerificationTokenRecord,
) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("upserting email verification token record failed")

	err := transaction.Get(record, `
		INSERT INTO email_verification_token_records (
			email,
			token_hash,
			expires_at,
			last_sent_at,
			hourly_send_count,
			hourly_window_started_at,
			daily_send_count,
			daily_window_started_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (email) DO UPDATE
		SET token_hash = EXCLUDED.token_hash,
		    expires_at = EXCLUDED.expires_at,
		    last_sent_at = EXCLUDED.last_sent_at,
		    hourly_send_count = EXCLUDED.hourly_send_count,
		    hourly_window_started_at = EXCLUDED.hourly_window_started_at,
		    daily_send_count = EXCLUDED.daily_send_count,
		    daily_window_started_at = EXCLUDED.daily_window_started_at
		RETURNING id
	`,
		record.Email,
		record.TokenHash,
		record.ExpiresAt,
		record.LastSentAt,
		record.HourlySendCount,
		record.HourlyWindowStartedAt,
		record.DailySendCount,
		record.DailyWindowStartedAt,
	)
	if err != nil {
		return wrapUnexpectedError(err)
	}

	return nil
}
