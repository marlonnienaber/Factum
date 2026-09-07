package services

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	categoriesStores "github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/domain"
	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"roci.dev/fracdex"
)

var (
	ErrTaskDateAndOrRankChangeNotAllowed = errors.New("requested date and/or rank change not allowed")
	ErrTaskAlreadyTrashed                = errors.New("task is already trashed")
	ErrTaskAlreadyRestored               = errors.New("task is already restored")
	ErrTaskAlreadyCompleted              = errors.New("task is already completed")
	ErrTaskAlreadyOpen                   = errors.New("task is already open")
	ErrTaskRankRequired                  = errors.New("task rank is required for an untrashed task")
)

type TaskService struct {
	db            *sqlx.DB
	taskStore     *stores.TaskStore
	categoryStore *categoriesStores.CategoryStore
}

func NewTaskService(
	db *sqlx.DB,
	taskStore *stores.TaskStore,
	categoryStore *categoriesStores.CategoryStore,
) *TaskService {
	return &TaskService{
		db:            db,
		taskStore:     taskStore,
		categoryStore: categoryStore,
	}
}

func (service *TaskService) CreateTask(userId int64, request transport.CreateTaskRequest) (*transport.TaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("creating task failed")

	task, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TaskResponse, error) {
		categoryExists, err := service.categoryStore.DoesCategoryExistForUser(transaction, request.CategoryId, userId)
		if err != nil {
			return nil, err
		}
		if !categoryExists {
			return nil, stores.ErrCategoryNotFound
		}

		var highestRank *string
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

		categoryExists, err := service.categoryStore.DoesCategoryExistForUser(transaction, request.CategoryId, userId)
		if err != nil {
			return nil, err
		}
		if !categoryExists {
			return nil, stores.ErrCategoryNotFound
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

		return task.ToTaskResponse(), nil
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

func (service *TaskService) ChangeTaskDateAndOrRank(userId int64, request transport.ChangeTaskDateAndOrRankRequest) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("changing task date and/or rank failed")

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, request.Id, userId)
		if err != nil {
			return err
		}

		// TODO: check once more whether the endpoint really only allows allowed transitions (map out transitions that UI allows and compare to logic)
		isMovingFromDateToDate := task.Date != nil && request.Date != nil && !datesAreEqual(task.Date, request.Date)
		isAllowedToMoveFromDateToDate := task.TrashedDate == nil && task.Date != nil

		isMovingFromDateToNoDate := task.Date != nil && request.Date == nil
		isAllowedToMoveFromDateToNoDate := task.TrashedDate == nil && task.Date != nil && !task.Done

		isMovingFromNoDateToDate := task.Date == nil && request.Date != nil
		isAllowedToMoveFromNoDateToDate := task.TrashedDate == nil && task.Date == nil && !task.Done

		isOnlyChangingRank := datesAreEqual(task.Date, request.Date)
		isAllowedToOnlyChangeRank := task.TrashedDate == nil

		isAllowedToChangeDateAndOrRank := isMovingFromDateToDate && isAllowedToMoveFromDateToDate ||
			isMovingFromDateToNoDate && isAllowedToMoveFromDateToNoDate ||
			isMovingFromNoDateToDate && isAllowedToMoveFromNoDateToDate ||
			isOnlyChangingRank && isAllowedToOnlyChangeRank

		if !isAllowedToChangeDateAndOrRank {
			return ErrTaskDateAndOrRankChangeNotAllowed
		}

		task.Date = request.Date
		task.Rank = &request.Rank
		if err := service.taskStore.UpdateTaskDateAndRank(transaction, task); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, stores.ErrTaskRankAlreadyExists),
			errors.Is(err, ErrTaskDateAndOrRankChangeNotAllowed):
			return err
		default:
			return wrapUnexpectedError(err)
		}
	}

	return nil
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

func (service *TaskService) BatchTrashTasks(userId int64, taskIds []int64, trashedDate time.Time) error {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch trashing tasks failed")

	err := shared.ExecuteAsSerializableTransactionWithoutResult(service.db, func(transaction *sqlx.Tx) error {
		tasks, err := service.taskStore.GetTasksByIdsForUser(transaction, taskIds, userId)
		if err != nil {
			return err
		}

		for index := range tasks {
			if tasks[index].TrashedDate != nil {
				return ErrTaskAlreadyTrashed
			}
		}

		return service.taskStore.TrashTasks(transaction, userId, taskIds, trashedDate)
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

func (service *TaskService) BatchCompleteTasks(userId int64, taskIds []int64) (*transport.BatchUpdateDoneResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch completing tasks failed")
	currentDate := shared.NewDate(time.Now())

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.BatchUpdateDoneResponse, error) {
		tasks, err := service.taskStore.GetTasksByIdsForUser(transaction, taskIds, userId)
		if err != nil {
			return nil, err
		}

		if err := service.checkThatTasksAreOpen(tasks); err != nil {
			return nil, err
		}
		service.setDoneFlagAndIfNeededDate(tasks, currentDate)

		untrashedTasks := service.extractUntrashedTasks(tasks)
		if err := service.assignRanksForBatchDoneUpdate(transaction, userId, true, untrashedTasks); err != nil {
			return nil, err
		}
		if err := service.taskStore.BatchUpdateDoneForUntrashedTasks(transaction, userId, untrashedTasks, true); err != nil {
			return nil, err
		}

		trashedTasks := service.extractTrashedTasks(tasks)
		if err := service.taskStore.BatchUpdateDoneForTrashedTasks(transaction, userId, trashedTasks, true); err != nil {
			return nil, err
		}

		return service.createBatchUpdateDoneResponse(tasks), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyCompleted):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) BatchOpenTasks(userId int64, taskIds []int64) (*transport.BatchUpdateDoneResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch opening tasks failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.BatchUpdateDoneResponse, error) {
		tasks, err := service.taskStore.GetTasksByIdsForUser(transaction, taskIds, userId)
		if err != nil {
			return nil, err
		}

		if err := service.checkThatTasksAreCompleted(tasks); err != nil {
			return nil, err
		}
		service.unsetDoneFlag(tasks)

		untrashedTasks := service.extractUntrashedTasks(tasks)
		if err := service.assignRanksForBatchDoneUpdate(transaction, userId, false, untrashedTasks); err != nil {
			return nil, err
		}
		if err := service.taskStore.BatchUpdateDoneForUntrashedTasks(transaction, userId, untrashedTasks, false); err != nil {
			return nil, err
		}

		trashedTasks := service.extractTrashedTasks(tasks)
		if err := service.taskStore.BatchUpdateDoneForTrashedTasks(transaction, userId, trashedTasks, false); err != nil {
			return nil, err
		}

		return service.createBatchUpdateDoneResponse(tasks), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyOpen):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) checkThatTasksAreOpen(tasks []domain.Task) error {
	for _, task := range tasks {
		if task.Done {
			return ErrTaskAlreadyCompleted
		}
	}
	return nil
}

func (service *TaskService) checkThatTasksAreCompleted(tasks []domain.Task) error {
	for _, task := range tasks {
		if !task.Done {
			return ErrTaskAlreadyOpen
		}
	}
	return nil
}

func (service *TaskService) setDoneFlagAndIfNeededDate(tasks []domain.Task, currentDate shared.Date) {
	for index := range tasks {
		tasks[index].Done = true
		if tasks[index].Date == nil {
			taskDate := currentDate
			tasks[index].Date = &taskDate
		}
	}
}

func (service *TaskService) unsetDoneFlag(tasks []domain.Task) {
	for index := range tasks {
		tasks[index].Done = false
	}
}

func (service *TaskService) assignRanksForBatchDoneUpdate(
	transaction *sqlx.Tx,
	userId int64,
	targetDoneState bool,
	untrashedTasks []*domain.Task,
) error {
	untrashedTaskDates := make([]shared.Date, 0)
	untrashedTasksByDate := make(map[shared.Date][]*domain.Task)
	for _, task := range untrashedTasks {
		taskDate := *task.Date
		if _, exists := untrashedTasksByDate[taskDate]; !exists {
			untrashedTaskDates = append(untrashedTaskDates, taskDate)
		}
		untrashedTasksByDate[taskDate] = append(untrashedTasksByDate[taskDate], task)
	}

	highestRanksByDate, err := service.taskStore.GetHighestUntrashedTaskRanksByDate(
		transaction,
		userId,
		untrashedTaskDates,
		targetDoneState,
	)
	if err != nil {
		return err
	}

	for _, taskDate := range untrashedTaskDates {
		lowerRankBoundary := highestRanksByDate[taskDate]
		for _, task := range untrashedTasksByDate[taskDate] {
			rank, err := fracdex.KeyBetween(lowerRankBoundary, "")
			if err != nil {
				return err
			}
			task.Rank = &rank
			lowerRankBoundary = rank
		}
	}
	return nil
}

func (service *TaskService) extractUntrashedTasks(tasks []domain.Task) []*domain.Task {
	untrashedTasks := make([]*domain.Task, 0, len(tasks)/2)
	for index := range tasks {
		if tasks[index].TrashedDate == nil {
			untrashedTasks = append(untrashedTasks, &tasks[index])
		}
	}
	return untrashedTasks
}

func (service *TaskService) extractTrashedTasks(tasks []domain.Task) []*domain.Task {
	trashedTasks := make([]*domain.Task, 0, len(tasks)/2)
	for index := range tasks {
		if tasks[index].TrashedDate != nil {
			trashedTasks = append(trashedTasks, &tasks[index])
		}
	}
	return trashedTasks
}

func (service *TaskService) createBatchUpdateDoneResponse(tasks []domain.Task) *transport.BatchUpdateDoneResponse {
	response := &transport.BatchUpdateDoneResponse{
		DatedUntrashedOpenTasks: make([]transport.TaskResponse, 0),
		DatedUntrashedDoneTasks: make([]transport.TaskResponse, 0),
		TrashedTasks:            make([]transport.TaskResponse, 0),
	}
	for index := range tasks {
		task := &tasks[index]
		taskResponse := *task.ToTaskResponse()
		switch {
		case task.TrashedDate != nil:
			response.TrashedTasks = append(response.TrashedTasks, taskResponse)
		case task.Done:
			response.DatedUntrashedDoneTasks = append(response.DatedUntrashedDoneTasks, taskResponse)
		default:
			response.DatedUntrashedOpenTasks = append(response.DatedUntrashedOpenTasks, taskResponse)
		}
	}
	return response
}

func datesAreEqual(firstDate *shared.Date, secondDate *shared.Date) bool {
	if firstDate == nil || secondDate == nil {
		return firstDate == nil && secondDate == nil
	}

	return firstDate.Time.Equal(secondDate.Time)
}
