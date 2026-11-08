package stores

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var (
	ErrEmailAlreadyInUse   = errors.New("the given email is already in use.")
	ErrNoUserForGivenEmail = errors.New("no user exists with the given email.")
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

func (store *UserStore) GetUserByEmail(email string) (*domain.User, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving user by email failed")

	var user domain.User

	err := store.db.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoUserForGivenEmail
		}
		return nil, wrapUnexpectedError(err)
	}

	return &user, nil
}

func (store *UserStore) DoesUserExistForId(userId int64) (bool, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("checking user existence by id failed")

	var exists bool

	err := store.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userId)
	if err != nil {
		return false, wrapUnexpectedError(err)
	}

	return exists, nil
}

func (store *UserStore) DoesUserExistForEmail(transaction *sqlx.Tx, email string) (bool, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("checking user existence by email failed")

	var exists bool
	err := transaction.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
	if err != nil {
		return false, wrapUnexpectedError(err)
	}

	return exists, nil
}
