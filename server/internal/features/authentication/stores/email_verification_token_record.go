package stores

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var ErrNoRecordForEmailVerificationToken = errors.New("no matching record for the given email verification token exists")

type EmailVerificationTokenRecordStore struct {
	db *sqlx.DB
}

func NewEmailVerificationTokenRecordStore(db *sqlx.DB) *EmailVerificationTokenRecordStore {
	return &EmailVerificationTokenRecordStore{
		db: db,
	}
}

func (store *EmailVerificationTokenRecordStore) GetRecordForEmail(
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
		       minutely_window_started_at,
		       hourly_send_count,
		       hourly_window_started_at,
		       daily_send_count,
		       daily_window_started_at
		FROM email_verification_token_records
		WHERE email = $1
	`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, wrapUnexpectedError(err)
	}

	return &record, nil
}

func (store *EmailVerificationTokenRecordStore) GetRecordForTokenHash(
	transaction *sqlx.Tx,
	tokenHash []byte,
) (*domain.EmailVerificationTokenRecord, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving email verification token record by token hash failed")

	var record domain.EmailVerificationTokenRecord
	err := transaction.Get(&record, `
		SELECT id,
		       email,
		       token_hash,
		       expires_at,
		       minutely_window_started_at,
		       hourly_send_count,
		       hourly_window_started_at,
		       daily_send_count,
		       daily_window_started_at
		FROM email_verification_token_records
		WHERE token_hash = $1
	`, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecordForEmailVerificationToken
		}
		return nil, wrapUnexpectedError(err)
	}

	return &record, nil
}

func (store *EmailVerificationTokenRecordStore) UpsertRecord(
	transaction *sqlx.Tx,
	record *domain.EmailVerificationTokenRecord,
) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("upserting email verification token record failed")

	err := transaction.Get(record, `
		INSERT INTO email_verification_token_records (
			email,
			token_hash,
			expires_at,
			minutely_window_started_at,
			hourly_send_count,
			hourly_window_started_at,
			daily_send_count,
			daily_window_started_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (email) DO UPDATE
		SET token_hash = EXCLUDED.token_hash,
		    expires_at = EXCLUDED.expires_at,
		    minutely_window_started_at = EXCLUDED.minutely_window_started_at,
		    hourly_send_count = EXCLUDED.hourly_send_count,
		    hourly_window_started_at = EXCLUDED.hourly_window_started_at,
		    daily_send_count = EXCLUDED.daily_send_count,
		    daily_window_started_at = EXCLUDED.daily_window_started_at
		RETURNING id
	`,
		record.Email,
		record.TokenHash,
		record.ExpiresAt,
		record.MinutelyWindowStartedAt,
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

func (store *EmailVerificationTokenRecordStore) DeleteRecord(
	transaction *sqlx.Tx,
	recordId int64,
) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("deleting email verification token record failed")

	result, err := transaction.Exec("DELETE FROM email_verification_token_records WHERE id = $1", recordId)
	if err != nil {
		return wrapUnexpectedError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapUnexpectedError(err)
	}
	if rowsAffected == 0 {
		return ErrNoRecordForEmailVerificationToken
	}

	return nil
}
