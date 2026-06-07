package stores

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var (
	ErrEmailAlreadyInUse = errors.New("the given email is already in use.")
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (store *UserStore) InsertUser(user *domain.User) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("inserting user failed")

	err := store.db.Get(user, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Constraint == "users_email_unique" {
			return ErrEmailAlreadyInUse
		}
		return wrapUnexpectedError(err)
	}
	return nil
}
