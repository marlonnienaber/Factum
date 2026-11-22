package services

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

var ErrMaximumCategoryCountReached = errors.New("maximum category count reached")

const maximumCategoryCountPerUser = 30

type CategoryService struct {
	db            *sqlx.DB
	categoryStore *stores.CategoryStore
}

func NewCategoryService(db *sqlx.DB, categoryStore *stores.CategoryStore) *CategoryService {
	return &CategoryService{
		db:            db,
		categoryStore: categoryStore,
	}
}

func (service *CategoryService) CreateCategory(
	ctx context.Context,
	userId int64,
	request transport.CreateCategoryRequest,
) (*transport.CreateCategoryResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("creating category failed")

	category, err := shared.ExecuteAsSerializableTransaction(ctx, service.db, func(transaction *sqlx.Tx) (*transport.CreateCategoryResponse, error) {
		categoryCount, err := service.categoryStore.CountCategoriesForUser(ctx, transaction, userId)
		if err != nil {
			return nil, err
		}
		if categoryCount >= maximumCategoryCountPerUser {
			return nil, ErrMaximumCategoryCountReached
		}

		return service.categoryStore.InsertCategory(ctx, transaction, userId, request.Title, request.Color)
	})
	if err != nil {
		if errors.Is(err, ErrMaximumCategoryCountReached) ||
			errors.Is(err, stores.ErrCategoryTitleAlreadyInUse) ||
			errors.Is(err, stores.ErrCategoryColorAlreadyInUse) {
			return nil, err
		}
		return nil, wrapUnexpectedError(err)
	}

	return category, nil
}
