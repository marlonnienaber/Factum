package stores

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

const tasksBatchSize = 50

var (
	ErrCategoryNotFound      = errors.New("no category exists for the given id")
	ErrTaskNotFound          = errors.New("no task exists for the given id and user id")
	ErrTaskRankAlreadyExists = errors.New("task rank already exists in the target task list")
)

type TaskStore struct {
	db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) *TaskStore {
	return &TaskStore{
		db: db,
	}
}

func (store *TaskStore) GetTaskByIdForUser(transaction *sqlx.Tx, id int64, userId int64) (*domain.Task, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving task by id and user id failed")

	var task domain.Task
	err := transaction.Get(&task, `
		SELECT id,
		       user_id,
		       category_id,
		       title,
		       description,
		       date,
		       rank,
		       done,
		       trashed_date
		FROM tasks
		WHERE id = $1
		  AND user_id = $2
	`, id, userId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return &task, nil
}

func (store *TaskStore) UpdateTaskDateAndRank(transaction *sqlx.Tx, task *domain.Task) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating task date and rank failed")

	result, err := transaction.Exec(`
		UPDATE tasks
		SET date = $1,
		    rank = $2
		WHERE id = $3
		  AND user_id = $4
	`, task.Date, task.Rank, task.Id, task.UserId)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Constraint == "tasks_untrashed_user_id_done_date_rank_unique" {
			return ErrTaskRankAlreadyExists
		}
		return wrapUnexpectedError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapUnexpectedError(err)
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (store *TaskStore) GetHighestUndatedUntrashedOpenTaskRank(transaction *sqlx.Tx, userId int64) (*string, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving highest undated untrashed undone task rank failed")

	var rank string
	err := transaction.Get(&rank, `
		SELECT rank
		FROM tasks
		WHERE user_id = $1
		  AND date IS NULL
		  AND trashed_date IS NULL
		  AND done = FALSE
		ORDER BY rank DESC
		LIMIT 1
	`, userId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return &rank, nil
}

func (store *TaskStore) GetHighestDatedUntrashedOpenTaskRank(transaction *sqlx.Tx, userId int64, date *shared.Date) (*string, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving highest dated untrashed open task rank failed")

	var rank string
	err := transaction.Get(&rank, `
		SELECT rank
		FROM tasks
		WHERE user_id = $1
		  AND date = $2
		  AND trashed_date IS NULL
		  AND done = FALSE
		ORDER BY rank DESC
		LIMIT 1
	`, userId, date)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return &rank, nil
}

func (store *TaskStore) InsertTask(
	transaction *sqlx.Tx,
	userId int64,
	request transport.CreateTaskRequest,
	rank string,
) (*transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("inserting task failed")

	var task transport.TaskResponse
	err := transaction.Get(&task, `
		INSERT INTO tasks (user_id, category_id, title, description, date, rank)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id,
		          category_id,
		          title,
		          description,
		          date,
		          rank,
		          done
	`,
		userId,
		request.CategoryId,
		request.Title,
		request.Description,
		request.Date,
		rank,
	)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) && postgresError.Constraint == "tasks_category_id_foreign_key" {
			return nil, ErrCategoryNotFound
		}
		return nil, wrapUnexpectedError(err)
	}

	return &task, nil
}

func (store *TaskStore) GetUndatedUntrashedUndoneTasks(userId int64, rank *string) ([]transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving undated untrashed undone tasks failed")

	tasks := make([]transport.TaskResponse, 0)
	err := store.db.Select(&tasks, `
		SELECT id,
		       category_id,
		       title,
		       description,
		       rank,
		       done
		FROM tasks
		WHERE user_id = $1
		  AND date IS NULL
		  AND trashed_date IS NULL
		  AND done = FALSE
		  AND ($2::TEXT IS NULL OR rank > $2)
		ORDER BY rank ASC
		LIMIT $3
	`, userId, rank, tasksBatchSize)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return tasks, nil
}

func (store *TaskStore) GetDatedUntrashedDoneTasks(userId int64, date *time.Time, rank *string) ([]transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving dated untrashed done tasks failed")

	tasks := make([]transport.TaskResponse, 0)
	err := store.db.Select(&tasks, `
		SELECT id,
		       category_id,
		       title,
		       description,
		       date,
		       rank,
		       done
		FROM tasks
		WHERE user_id = $1
		  AND date IS NOT NULL
		  AND trashed_date IS NULL
		  AND done = TRUE
		  AND (
		      ($2::DATE IS NULL AND $3::TEXT IS NULL)
		      OR (
		          $2::DATE IS NOT NULL
		          AND $3::TEXT IS NOT NULL
		          AND (
		              date > $2
		              OR (date = $2 AND rank > $3)
		          )
		      )
		  )
		ORDER BY date ASC, rank ASC
		LIMIT $4
	`, userId, date, rank, tasksBatchSize)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return tasks, nil
}

func (store *TaskStore) GetDatedUntrashedOpenTasks(userId int64, date *time.Time, rank *string) ([]transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving dated untrashed undone tasks failed")

	tasks := make([]transport.TaskResponse, 0)
	err := store.db.Select(&tasks, `
		SELECT id,
		       category_id,
		       title,
		       description,
		       date,
		       rank,
		       done
		FROM tasks
		WHERE user_id = $1
		  AND date IS NOT NULL
		  AND trashed_date IS NULL
		  AND done = FALSE
		  AND (
		      ($2::DATE IS NULL AND $3::TEXT IS NULL)
		      OR (
		          $2::DATE IS NOT NULL
		          AND $3::TEXT IS NOT NULL
		          AND (
		              date > $2
		              OR (date = $2 AND rank > $3)
		          )
		      )
		  )
		ORDER BY date ASC, rank ASC
		LIMIT $4
	`, userId, date, rank, tasksBatchSize)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return tasks, nil
}

func (store *TaskStore) GetTrashedTasks(userId int64, trashedTimestamp *time.Time, id *int64) ([]transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving trashed tasks failed")

	tasks := make([]transport.TaskResponse, 0)
	err := store.db.Select(&tasks, `
		SELECT id,
		       category_id,
		       title,
		       description,
		       date,
		       rank,
		       done,
		       trashed_date
		FROM tasks
		WHERE user_id = $1
		  AND trashed_date IS NOT NULL
		  AND (
		      ($2::TIMESTAMPTZ IS NULL AND $3::BIGINT IS NULL)
		      OR (
		          $2::TIMESTAMPTZ IS NOT NULL
		          AND $3::BIGINT IS NOT NULL
		          AND (
		              trashed_date < $2
		              OR (trashed_date = $2 AND id < $3)
		          )
		      )
		  )
		ORDER BY trashed_date DESC, id DESC
		LIMIT $4
	`, userId, trashedTimestamp, id, tasksBatchSize)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	return tasks, nil
}
