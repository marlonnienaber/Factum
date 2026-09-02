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

type dateAndRankPair struct {
	Date shared.Date `db:"date"`
	Rank string      `db:"rank"`
}

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

func (store *TaskStore) GetTasksByIdsForUser(transaction *sqlx.Tx, ids []int64, userId int64) ([]domain.Task, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving tasks by ids and user id failed")

	tasks := make([]domain.Task, 0, len(ids))
	err := transaction.Select(&tasks, `
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
		WHERE id = ANY($1::BIGINT[])
		  AND user_id = $2
	`, pq.Array(ids), userId)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}
	if len(tasks) != len(ids) {
		return nil, ErrTaskNotFound
	}

	return tasks, nil
}

func (store *TaskStore) GetHighestUntrashedTaskRanksByDate(
	transaction *sqlx.Tx,
	userId int64,
	dates []shared.Date,
	done bool,
) (map[shared.Date]string, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving highest untrashed task ranks by date failed")

	if len(dates) == 0 {
		return map[shared.Date]string{}, nil
	}

	var dateAndRankPairs []dateAndRankPair
	err := transaction.Select(&dateAndRankPairs, `
		SELECT date,
		       MAX(rank) AS rank
		FROM tasks
		WHERE user_id = $1
		  AND date = ANY($2::DATE[])
		  AND done = $3
		  AND trashed_date IS NULL
		GROUP BY date
	`, userId, pq.Array(dates), done)
	if err != nil {
		return nil, wrapUnexpectedError(err)
	}

	highestRanksByDate := make(map[shared.Date]string, len(dateAndRankPairs))
	for _, dateWithHighestRank := range dateAndRankPairs {
		highestRanksByDate[dateWithHighestRank.Date] = dateWithHighestRank.Rank
	}
	return highestRanksByDate, nil
}

func (store *TaskStore) UpdateUntrashedTasksDoneState(
	transaction *sqlx.Tx,
	userId int64,
	tasks []*domain.Task,
	targetDoneState bool,
) error {
	if len(tasks) == 0 {
		return nil
	}

	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating untrashed tasks done state failed")

	ids := make([]int64, 0, len(tasks))
	dates := make([]shared.Date, 0, len(tasks))
	ranks := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.Id)
		dates = append(dates, *task.Date)
		ranks = append(ranks, *task.Rank)
	}

	result, err := transaction.Exec(`
		UPDATE tasks
		SET date = updates.date,
		    rank = updates.rank,
		    done = $4
		FROM UNNEST(
			$1::BIGINT[],
			$2::DATE[],
			$3::TEXT[]
		) AS updates(id, date, rank)
		WHERE tasks.id = updates.id
		  AND tasks.user_id = $5
		  AND tasks.trashed_date IS NULL
	`, pq.Array(ids), pq.Array(dates), pq.Array(ranks), targetDoneState, userId)
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
	if rowsAffected != int64(len(tasks)) {
		return ErrTaskNotFound
	}
	return nil
}

func (store *TaskStore) UpdateTrashedTasksDoneState(
	transaction *sqlx.Tx,
	userId int64,
	tasks []*domain.Task,
	targetDoneState bool,
) error {
	if len(tasks) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(tasks))
	dates := make([]shared.Date, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.Id)
		dates = append(dates, *task.Date)
	}

	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating trashed tasks done state failed")
	result, err := transaction.Exec(`
		UPDATE tasks
		SET date = updates.date,
		    done = $3
		FROM UNNEST(
			$1::BIGINT[],
			$2::DATE[]
		) AS updates(id, date)
		WHERE tasks.id = updates.id
		  AND tasks.user_id = $4
		  AND tasks.trashed_date IS NOT NULL
	`, pq.Array(ids), pq.Array(dates), targetDoneState, userId)
	if err != nil {
		return wrapUnexpectedError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapUnexpectedError(err)
	}
	if rowsAffected != int64(len(tasks)) {
		return ErrTaskNotFound
	}
	return nil
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

func (store *TaskStore) UpdateTaskTrashedState(transaction *sqlx.Tx, task *domain.Task) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating task trash state failed")

	result, err := transaction.Exec(`
		UPDATE tasks
		SET trashed_date = $1,
		    rank = $2
		WHERE id = $3
		  AND user_id = $4
	`, task.TrashedDate, task.Rank, task.Id, task.UserId)
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

func (store *TaskStore) UpdateTaskDoneState(transaction *sqlx.Tx, task *domain.Task) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating task done state failed")

	result, err := transaction.Exec(`
		UPDATE tasks
		SET date = $1,
		    rank = $2,
		    done = $3
		WHERE id = $4
		  AND user_id = $5
	`, task.Date, task.Rank, task.Done, task.Id, task.UserId)
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

func (store *TaskStore) UpdateTask(transaction *sqlx.Tx, task *domain.Task) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating task failed")

	result, err := transaction.Exec(`
		UPDATE tasks
		SET category_id = $1,
		    title = $2,
		    description = $3,
		    date = $4,
		    rank = $5,
		    done = $6
		WHERE id = $7
		  AND user_id = $8
	`,
		task.CategoryId,
		task.Title,
		task.Description,
		task.Date,
		task.Rank,
		task.Done,
		task.Id,
		task.UserId,
	)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) {
			switch postgresError.Constraint {
			case "tasks_category_id_foreign_key":
				return ErrCategoryNotFound
			case "tasks_untrashed_user_id_done_date_rank_unique":
				return ErrTaskRankAlreadyExists
			}
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

func (store *TaskStore) GetHighestDatedUntrashedDoneTaskRank(transaction *sqlx.Tx, userId int64, date *shared.Date) (*string, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving highest dated untrashed done task rank failed")

	var rank string
	err := transaction.Get(&rank, `
		SELECT rank
		FROM tasks
		WHERE user_id = $1
		  AND date = $2
		  AND trashed_date IS NULL
		  AND done = TRUE
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

func (store *TaskStore) GetUndatedUntrashedUndoneTasks(userId int64, rank *string) ([]transport.UndatedTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("retrieving undated untrashed undone tasks failed")

	tasks := make([]transport.UndatedTaskResponse, 0)
	err := store.db.Select(&tasks, `
		SELECT id,
		       category_id,
		       title,
		       description,
		       rank
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

func (store *TaskStore) GetDatedDoneUntrashedTasks(userId int64, date *time.Time, rank *string) ([]transport.TaskResponse, error) {
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

func (store *TaskStore) GetDatedOpenUntrashedTasks(userId int64, date *time.Time, rank *string) ([]transport.TaskResponse, error) {
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
