package domain

import "time"

type EmailVerificationTokenRecord struct {
	Id                      int64     `db:"id"`
	Email                   string    `db:"email"`
	TokenHash               []byte    `db:"token_hash"`
	ExpiresAt               time.Time `db:"expires_at"`
	MinutelyWindowStartedAt time.Time `db:"minutely_window_started_at"`
	HourlySendCount         int       `db:"hourly_send_count"`
	HourlyWindowStartedAt   time.Time `db:"hourly_window_started_at"`
	DailySendCount          int       `db:"daily_send_count"`
	DailyWindowStartedAt    time.Time `db:"daily_window_started_at"`
}
