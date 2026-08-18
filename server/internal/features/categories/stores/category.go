package stores

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var (
	ErrCategoryNotFound          = errors.New("no category exists for the given id and user id")
	ErrCategoryStillReferenced   = errors.New("category is still referenced by tasks")
	ErrCategoryTitleAlreadyInUse = errors.New("the given category title is already in use for the user")
)

type CategoryStore struct {
	db *sqlx.DB
}

func NewCategoryStore(db *sqlx.DB) *CategoryStore {
	return &CategoryStore{
		db: db,
	}
}

func (store *CategoryStore) GetCategories(userId int64) ([]transport.CategoryResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving categories failed")

	categories := make([]transport.CategoryResponse, 0)
	err := store.db.Select(&categories, `
		SELECT id,
		       title
		FROM categories
		WHERE user_id = $1
		ORDER BY id ASC
	`, userId)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return categories, nil
}

func (store *CategoryStore) DoesCategoryExistForUser(
	transaction *sqlx.Tx,
	id int64,
	userId int64,
) (bool, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("checking category existence for user failed")

	var exists bool
	err := transaction.Get(&exists, `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE id = $1
			  AND user_id = $2
		)
	`, id, userId)
	if err != nil {
		return false, wrapUnexpectedError(err)
	}

	return exists, nil
}

func (store *CategoryStore) InsertCategory(userId int64, title string) (*transport.CategoryResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("inserting category failed")

	var category transport.CategoryResponse
	err := store.db.Get(&category, `
		INSERT INTO categories (user_id, title)
		VALUES ($1, $2)
		RETURNING id, title
	`, userId, title)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Constraint == "categories_user_id_title_unique" {
			return nil, ErrCategoryTitleAlreadyInUse
		}
		return nil, wrapUnexpectedError(err)
	}

	return &category, nil
}

func (store *CategoryStore) DeleteCategory(id int64, userId int64) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("deleting category failed")

	result, err := store.db.Exec("DELETE FROM categories WHERE id = $1 AND user_id = $2", id, userId)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Constraint == "tasks_category_id_foreign_key" {
			return ErrCategoryStillReferenced
		}
		return wrapUnexpectedError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapUnexpectedError(err)
	}
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
