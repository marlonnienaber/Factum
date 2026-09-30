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
	ErrRequestDataRepresentInvalidDragAndDropGesture = errors.New("provided request data do not match a valid drag and drop gesture for the given task id")
	ErrPreviousTaskNotFound                          = errors.New("previous task not found")
	ErrMaximumTaskCountReached                       = errors.New("maximum task count reached")
	ErrTaskAlreadyTrashed                            = errors.New("task is already trashed")
	ErrTaskIsTrashed                                 = errors.New("task is trashed")
	ErrTaskAlreadyRestored                           = errors.New("task is already restored")
	ErrTaskAlreadyCompleted                          = errors.New("task is already completed")
	ErrTaskAlreadyOpen                               = errors.New("task is already open")
)

const maximumTaskCountPerUser = 1_000

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

func (service *TaskService) CreateTask(userId int64, request transport.CreateTaskRequest) (*transport.CreateTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("creating task failed")

	task, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.CreateTaskResponse, error) {
		taskCount, err := service.taskStore.CountTasksForUser(transaction, userId)
		if err != nil {
			return nil, err
		}
		if taskCount >= maximumTaskCountPerUser {
			return nil, ErrMaximumTaskCountReached
		}

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
		if errors.Is(err, stores.ErrCategoryNotFound) || errors.Is(err, ErrMaximumTaskCountReached) {
			return nil, err
		}
		return nil, wrapUnexpectedError(err)
	}

	return task, nil
}

func (service *TaskService) UpdateTask(userId int64, request transport.UpdateTaskRequest) (*transport.UpdateTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("updating task failed")

	taskResponse, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.UpdateTaskResponse, error) {
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

		return task.ToUpdateTaskResponse(), nil
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

func (service *TaskService) DragAndDropTask(userId int64, taskId int64, request transport.TaskDragAndDropRequest) (*transport.TaskDragAndDropResponse, error) {
	// TODO: extend to support backlog and workbench cases
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("dragging and dropping task failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TaskDragAndDropResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return nil, err
		}

		isMovingFromDateToDate := task.Date != nil && request.Date != nil && !datesAreEqual(task.Date, request.Date)
		isAllowedToMoveFromDateToDate := task.TrashedDate == nil && !task.Done && task.Date != nil

		isMovingFromDateToNoDate := task.Date != nil && request.Date == nil
		isAllowedToMoveFromDateToNoDate := task.TrashedDate == nil && !task.Done && task.Date != nil

		isMovingFromNoDateToDate := task.Date == nil && request.Date != nil
		isAllowedToMoveFromNoDateToDate := task.TrashedDate == nil && !task.Done && task.Date == nil

		isOnlyChangingRank := datesAreEqual(task.Date, request.Date)
		isAllowedToOnlyChangeRank := task.TrashedDate == nil && !task.Done

		isAllowedToChangeDateAndOrRank := isMovingFromDateToDate && isAllowedToMoveFromDateToDate ||
			isMovingFromDateToNoDate && isAllowedToMoveFromDateToNoDate ||
			isMovingFromNoDateToDate && isAllowedToMoveFromNoDateToDate ||
			isOnlyChangingRank && isAllowedToOnlyChangeRank

		if !isAllowedToChangeDateAndOrRank {
			return nil, ErrRequestDataRepresentInvalidDragAndDropGesture
		}

		previousTaskRank := ""
		if request.PreviousTaskId != nil {
			if *request.PreviousTaskId == taskId {
				return nil, ErrRequestDataRepresentInvalidDragAndDropGesture
			}
			previousTask, err := service.taskStore.GetTaskByIdForUser(transaction, *request.PreviousTaskId, userId)
			if err != nil {
				if errors.Is(err, stores.ErrTaskNotFound) {
					return nil, ErrPreviousTaskNotFound
				}
				return nil, err
			}
			if !datesAreEqual(previousTask.Date, request.Date) || previousTask.Done != task.Done || previousTask.TrashedDate != nil || previousTask.Rank == nil {
				return nil, ErrRequestDataRepresentInvalidDragAndDropGesture
			}
			previousTaskRank = *previousTask.Rank
		}

		successorTaskRank, err := service.taskStore.GetRankOfSuccessorTaskForTaskDragAndDrop(transaction, userId, taskId, request.Date, task.Done, previousTaskRank)
		if err != nil {
			return nil, err
		}
		rank, err := fracdex.KeyBetween(previousTaskRank, successorTaskRank)
		if err != nil {
			return nil, err
		}
		task.Date = request.Date
		task.Rank = &rank
		if err := service.taskStore.UpdateTaskDateAndRank(transaction, task); err != nil {
			return nil, err
		}
		return &transport.TaskDragAndDropResponse{Rank: rank}, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrRequestDataRepresentInvalidDragAndDropGesture),
			errors.Is(err, ErrPreviousTaskNotFound):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}
	return response, nil
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

func (service *TaskService) RestoreTask(userId int64, taskId int64) (*transport.RestoreTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("restoring task failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.RestoreTaskResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return nil, err
		}
		if task.TrashedDate == nil {
			return nil, ErrTaskAlreadyRestored
		}

		var highestRank *string
		switch {
		case task.Date == nil:
			highestRank, err = service.taskStore.GetHighestUndatedUntrashedOpenTaskRank(transaction, userId)
		case task.Done:
			highestRank, err = service.taskStore.GetHighestDatedUntrashedDoneTaskRank(transaction, userId, task.Date)
		default:
			highestRank, err = service.taskStore.GetHighestDatedUntrashedOpenTaskRank(transaction, userId, task.Date)
		}
		if err != nil {
			return nil, err
		}

		previousTaskRank := ""
		if highestRank != nil {
			previousTaskRank = *highestRank
		}
		rank, err := fracdex.KeyBetween(previousTaskRank, "")
		if err != nil {
			return nil, err
		}
		task.TrashedDate = nil
		task.Rank = &rank
		if err := service.taskStore.UpdateTaskTrashedState(transaction, task); err != nil {
			return nil, err
		}

		return &transport.RestoreTaskResponse{Rank: rank}, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyRestored):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) ToggleTaskDone(userId int64, taskId int64) (*transport.ToggleTaskDoneResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("toggling task done failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.ToggleTaskDoneResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return nil, err
		}
		task.Done = !task.Done
		if task.Done && task.Date == nil {
			currentDate := shared.NewDate(time.Now())
			task.Date = &currentDate
		}
		if task.TrashedDate != nil {
			if err := service.taskStore.UpdateTaskDoneState(transaction, task); err != nil {
				return nil, err
			}
			return &transport.ToggleTaskDoneResponse{Date: *task.Date}, nil
		}

		var highestRank *string
		if task.Done {
			highestRank, err = service.taskStore.GetHighestDatedUntrashedDoneTaskRank(transaction, userId, task.Date)
		} else {
			highestRank, err = service.taskStore.GetHighestDatedUntrashedOpenTaskRank(transaction, userId, task.Date)
		}
		if err != nil {
			return nil, err
		}
		previousTaskRank := ""
		if highestRank != nil {
			previousTaskRank = *highestRank
		}
		rank, err := fracdex.KeyBetween(previousTaskRank, "")
		if err != nil {
			return nil, err
		}
		task.Rank = &rank

		if err := service.taskStore.UpdateTaskDoneState(transaction, task); err != nil {
			return nil, err
		}
		return &transport.ToggleTaskDoneResponse{Date: *task.Date, Rank: &rank}, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) BatchCompleteTasks(userId int64, taskIds []int64) ([]transport.BatchCompleteTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch completing tasks failed")
	currentDate := shared.NewDate(time.Now())

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) ([]transport.BatchCompleteTaskResponse, error) {
		tasks, err := service.taskStore.GetTasksByIdsForUser(transaction, taskIds, userId)
		if err != nil {
			return nil, err
		}

		if err := service.checkThatTasksAreUntrashed(tasks); err != nil {
			return nil, err
		}
		if err := service.checkThatTasksAreOpen(tasks); err != nil {
			return nil, err
		}
		service.setDoneFlagAndIfNeededDate(tasks, currentDate)

		if err := service.assignRanksForBatchDoneUpdate(transaction, userId, true, tasks); err != nil {
			return nil, err
		}
		if err := service.taskStore.BatchUpdateDoneForUntrashedTasks(transaction, userId, tasks, true); err != nil {
			return nil, err
		}

		return service.createBatchCompleteTasksResponse(tasks), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyCompleted),
			errors.Is(err, ErrTaskIsTrashed):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) BatchOpenTasks(userId int64, taskIds []int64) ([]transport.BatchOpenTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch opening tasks failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) ([]transport.BatchOpenTaskResponse, error) {
		tasks, err := service.taskStore.GetTasksByIdsForUser(transaction, taskIds, userId)
		if err != nil {
			return nil, err
		}

		if err := service.checkThatTasksAreUntrashed(tasks); err != nil {
			return nil, err
		}
		if err := service.checkThatTasksAreCompleted(tasks); err != nil {
			return nil, err
		}
		service.unsetDoneFlag(tasks)

		if err := service.assignRanksForBatchDoneUpdate(transaction, userId, false, tasks); err != nil {
			return nil, err
		}
		if err := service.taskStore.BatchUpdateDoneForUntrashedTasks(transaction, userId, tasks, false); err != nil {
			return nil, err
		}

		return service.createBatchOpenTasksResponse(tasks), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyOpen),
			errors.Is(err, ErrTaskIsTrashed):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) checkThatTasksAreOpen(tasks []*domain.Task) error {
	for _, task := range tasks {
		if task.Done {
			return ErrTaskAlreadyCompleted
		}
	}
	return nil
}

func (service *TaskService) checkThatTasksAreCompleted(tasks []*domain.Task) error {
	for _, task := range tasks {
		if !task.Done {
			return ErrTaskAlreadyOpen
		}
	}
	return nil
}

func (service *TaskService) checkThatTasksAreUntrashed(tasks []*domain.Task) error {
	for _, task := range tasks {
		if task.TrashedDate != nil {
			return ErrTaskIsTrashed
		}
	}
	return nil
}

func (service *TaskService) setDoneFlagAndIfNeededDate(tasks []*domain.Task, currentDate shared.Date) {
	for _, task := range tasks {
		task.Done = true
		if task.Date == nil {
			taskDate := currentDate
			task.Date = &taskDate
		}
	}
}

func (service *TaskService) unsetDoneFlag(tasks []*domain.Task) {
	for _, task := range tasks {
		task.Done = false
	}
}

func (service *TaskService) assignRanksForBatchDoneUpdate(
	transaction *sqlx.Tx,
	userId int64,
	targetDoneState bool,
	tasks []*domain.Task,
) error {
	taskDates := make([]shared.Date, 0)
	tasksByDate := make(map[shared.Date][]*domain.Task)
	for _, task := range tasks {
		taskDate := *task.Date
		if _, exists := tasksByDate[taskDate]; !exists {
			taskDates = append(taskDates, taskDate)
		}
		tasksByDate[taskDate] = append(tasksByDate[taskDate], task)
	}

	highestRanksByDate, err := service.taskStore.GetHighestUntrashedTaskRanksByDate(
		transaction,
		userId,
		taskDates,
		targetDoneState,
	)
	if err != nil {
		return err
	}

	for _, taskDate := range taskDates {
		lowerRankBoundary := highestRanksByDate[taskDate]
		for _, task := range tasksByDate[taskDate] {
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

func (service *TaskService) createBatchCompleteTasksResponse(tasks []*domain.Task) []transport.BatchCompleteTaskResponse {
	response := make([]transport.BatchCompleteTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, transport.BatchCompleteTaskResponse{
			Id:   task.Id,
			Date: *task.Date,
			Rank: *task.Rank,
		})
	}
	return response
}

func (service *TaskService) createBatchOpenTasksResponse(tasks []*domain.Task) []transport.BatchOpenTaskResponse {
	response := make([]transport.BatchOpenTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, transport.BatchOpenTaskResponse{
			Id:   task.Id,
			Rank: *task.Rank,
		})
	}
	return response
}

func datesAreEqual(firstDate *shared.Date, secondDate *shared.Date) bool {
	if firstDate == nil || secondDate == nil {
		return firstDate == nil && secondDate == nil
	}

	return firstDate.Time.Equal(secondDate.Time)
}
