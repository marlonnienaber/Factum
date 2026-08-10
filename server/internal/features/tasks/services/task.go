package services

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"roci.dev/fracdex"
)

var (
	ErrTaskDateChangeNotAllowed = errors.New("task cannot transition between its current and requested date states")
	ErrTaskAlreadyTrashed       = errors.New("task is already trashed")
	ErrTaskAlreadyRestored      = errors.New("task is already restored")
	ErrTaskRankRequired         = errors.New("task rank is required for an untrashed task")
)

type TaskService struct {
	db        *sqlx.DB
	taskStore *stores.TaskStore
}

func NewTaskService(db *sqlx.DB, taskStore *stores.TaskStore) *TaskService {
	return &TaskService{
		db:        db,
		taskStore: taskStore,
	}
}

func (service *TaskService) CreateTask(userId int64, request transport.CreateTaskRequest) (*transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("creating task failed")

	task, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TaskResponse, error) {
		var highestRank *string
		var err error
		if request.Date == nil {
			highestRank, err = service.taskStore.GetHighestUndatedUntrashedOpenTaskRank(transaction, userId)
		} else {
			highestRank, err = service.taskStore.GetHighestDatedUntrashedOpenTaskRank(transaction, userId, request.Date)
		}
		if err != nil {
			return nil, err
		}

		lowerRankBoundary := ""
		if highestRank != nil {
			lowerRankBoundary = *highestRank
		}
		rank, err := fracdex.KeyBetween(lowerRankBoundary, "")
		if err != nil {
			return nil, err
		}

		return service.taskStore.InsertTask(transaction, userId, request, rank)
	})
	if err != nil {
		if errors.Is(err, stores.ErrCategoryNotFound) {
			return nil, err
		}
		return nil, wrapUnexpectedError(err)
	}

	return task, nil
}

func (service *TaskService) UpdateTask(userId int64, request transport.UpdateTaskRequest) (*transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating task failed")

	taskResponse, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TaskResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, request.Id, userId)
		if err != nil {
			return nil, err
		}

		updatedDone := task.Done
		if request.Date == nil {
			updatedDone = false
		}

		dateChanged := !datesAreEqual(task.Date, request.Date)
		targetPartitionChanged := task.TrashedDate == nil && dateChanged
		if targetPartitionChanged {
			var highestRank *string
			switch {
			case request.Date == nil:
				highestRank, err = service.taskStore.GetHighestUndatedUntrashedOpenTaskRank(transaction, userId)
			case updatedDone:
				highestRank, err = service.taskStore.GetHighestDatedUntrashedDoneTaskRank(transaction, userId, request.Date)
			default:
				highestRank, err = service.taskStore.GetHighestDatedUntrashedOpenTaskRank(transaction, userId, request.Date)
			}
			if err != nil {
				return nil, err
			}

			lowerRankBoundary := ""
			if highestRank != nil {
				lowerRankBoundary = *highestRank
			}
			rank, err := fracdex.KeyBetween(lowerRankBoundary, "")
			if err != nil {
				return nil, err
			}
			task.Rank = &rank
		}

		task.CategoryId = request.CategoryId
		task.Title = request.Title
		task.Description = request.Description
		task.Date = request.Date
		task.Done = updatedDone

		if err := service.taskStore.UpdateTask(transaction, task); err != nil {
			return nil, err
		}

		return newTaskResponse(task), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, stores.ErrCategoryNotFound):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return taskResponse, nil
}

func (service *TaskService) ChangeTaskDate(userId int64, request transport.ChangeTaskDateRequest) (*transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("changing task date failed")

	taskResponse, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TaskResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, request.Id, userId)
		if err != nil {
			return nil, err
		}

		isMovingFromUndatedToDated := task.Date == nil && request.Date != nil
		isMovingFromDatedToUndated := task.Date != nil && request.Date == nil
		if task.Done || task.TrashedDate != nil || (!isMovingFromUndatedToDated && !isMovingFromDatedToUndated) {
			return nil, ErrTaskDateChangeNotAllowed
		}

		task.Date = request.Date
		task.Rank = &request.Rank
		if err := service.taskStore.UpdateTaskDateAndRank(transaction, task); err != nil {
			return nil, err
		}

		return newTaskResponse(task), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, stores.ErrTaskRankAlreadyExists),
			errors.Is(err, ErrTaskDateChangeNotAllowed):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return taskResponse, nil
}

func (service *TaskService) TrashTask(userId int64, taskId int64, trashedDate time.Time) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("trashing task failed")

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return err
		}
		if task.TrashedDate != nil {
			return ErrTaskAlreadyTrashed
		}

		task.TrashedDate = &trashedDate
		task.Rank = nil
		return service.taskStore.UpdateTaskTrashedState(transaction, task)
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyTrashed):
			return err
		default:
			return wrapUnexpectedError(err)
		}
	}

	return nil
}

func (service *TaskService) RestoreTask(userId int64, taskId int64, rank string) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("restoring task failed")

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return err
		}
		if task.TrashedDate == nil {
			return ErrTaskAlreadyRestored
		}

		task.TrashedDate = nil
		task.Rank = &rank
		return service.taskStore.UpdateTaskTrashedState(transaction, task)
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, stores.ErrTaskRankAlreadyExists),
			errors.Is(err, ErrTaskAlreadyRestored):
			return err
		default:
			return wrapUnexpectedError(err)
		}
	}

	return nil
}

func (service *TaskService) ToggleTaskDone(userId int64, taskId int64, rank *string) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("toggling task done failed")

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return err
		}
		if task.TrashedDate == nil {
			if rank == nil {
				return ErrTaskRankRequired
			}
			task.Rank = rank
		}

		task.Done = !task.Done
		if task.Done && task.Date == nil {
			currentDate := shared.NewDate(time.Now())
			task.Date = &currentDate
		}

		return service.taskStore.UpdateTaskDoneState(transaction, task)
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, stores.ErrTaskRankAlreadyExists),
			errors.Is(err, ErrTaskRankRequired):
			return err
		default:
			return wrapUnexpectedError(err)
		}
	}

	return nil
}

func newTaskResponse(task *domain.Task) *transport.TaskResponse {
	return &transport.TaskResponse{
		Id:          task.Id,
		CategoryId:  task.CategoryId,
		Title:       task.Title,
		Description: task.Description,
		Date:        task.Date,
		Rank:        task.Rank,
		Done:        task.Done,
		TrashedDate: task.TrashedDate,
	}
}

func datesAreEqual(firstDate *shared.Date, secondDate *shared.Date) bool {
	if firstDate == nil || secondDate == nil {
		return firstDate == nil && secondDate == nil
	}

	return firstDate.Time.Equal(secondDate.Time)
}
