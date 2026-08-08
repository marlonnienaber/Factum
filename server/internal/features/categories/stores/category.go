package stores

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var (
	ErrCategoryNotFound        = errors.New("no category exists for the given id")
	ErrCategoryStillReferenced = errors.New("category is still referenced by tasks")
)

type CategoryStore struct {
	db *sqlx.DB
}

func NewCategoryStore(db *sqlx.DB) *CategoryStore {
	return &CategoryStore{
		db: db,
	}
}

func (store *CategoryStore) InsertCategory(title string) (*transport.CategoryResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("inserting category failed")

	var category transport.CategoryResponse
	err := store.db.Get(&category, `
		INSERT INTO categories (title)
		VALUES ($1)
		RETURNING id, title
	`, title)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return &category, nil
}

func (store *CategoryStore) DeleteCategory(id int64) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("deleting category failed")

	result, err := store.db.Exec("DELETE FROM categories WHERE id = $1", id)
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
