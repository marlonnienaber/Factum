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
		targetPartitionChanged := !task.Trashed && dateChanged
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
			task.Rank = rank
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
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("dragging and dropping task failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TaskDragAndDropResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return nil, err
		}

		isValidDragAndDropAction := service.isValidDragAndDropAction(task, request)
		if !isValidDragAndDropAction {
			return nil, ErrRequestDataRepresentInvalidDragAndDropGesture
		}

		previousTaskRank, err := service.getPreviousTaskRankForTaskDragAndDrop(transaction, userId, taskId, request)
		if err != nil {
			return nil, err
		}
		successorTaskRank, err := service.taskStore.GetRankOfSuccessorTaskForTaskDragAndDrop(
			transaction,
			userId,
			taskId,
			request.Date,
			request.Done,
			request.Trashed,
			previousTaskRank,
		)
		if err != nil {
			return nil, err
		}
		rank, err := fracdex.KeyBetween(previousTaskRank, successorTaskRank)
		if err != nil {
			return nil, err
		}

		task.Date = request.Date
		task.Done = request.Done
		task.Trashed = request.Trashed
		task.Rank = rank
		if err := service.taskStore.UpdateTaskForDragAndDrop(transaction, task); err != nil {
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

func (service *TaskService) TrashTask(userId int64, taskId int64) (*transport.TrashTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("trashing task failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.TrashTaskResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return nil, err
		}
		if task.Trashed {
			return nil, ErrTaskAlreadyTrashed
		}

		lowestTrashedTaskRank, err := service.taskStore.GetLowestTrashedTaskRank(transaction, userId)
		if err != nil {
			return nil, err
		}
		upperRankBoundary := ""
		if lowestTrashedTaskRank != nil {
			upperRankBoundary = *lowestTrashedTaskRank
		}
		rank, err := fracdex.KeyBetween("", upperRankBoundary)
		if err != nil {
			return nil, err
		}

		if err := service.taskStore.TrashTask(transaction, userId, taskId, rank); err != nil {
			return nil, err
		}
		return &transport.TrashTaskResponse{Rank: rank}, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyTrashed):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) BatchTrashTasks(userId int64, taskIds []int64) ([]transport.BatchTrashTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch trashing tasks failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) ([]transport.BatchTrashTaskResponse, error) {
		tasks, err := service.taskStore.GetTasksByIdsForUser(transaction, taskIds, userId)
		if err != nil {
			return nil, err
		}

		for _, task := range tasks {
			if task.Trashed {
				return nil, ErrTaskAlreadyTrashed
			}
		}

		lowestTrashedTaskRank, err := service.taskStore.GetLowestTrashedTaskRank(transaction, userId)
		if err != nil {
			return nil, err
		}

		upperRankBoundary := ""
		if lowestTrashedTaskRank != nil {
			upperRankBoundary = *lowestTrashedTaskRank
		}
		ranks := make([]string, len(taskIds))
		for index := len(taskIds) - 1; index >= 0; index-- {
			rank, err := fracdex.KeyBetween("", upperRankBoundary)
			if err != nil {
				return nil, err
			}
			ranks[index] = rank
			upperRankBoundary = rank
		}

		if err := service.taskStore.TrashTasks(transaction, userId, taskIds, ranks); err != nil {
			return nil, err
		}

		return service.createBatchTrashTasksResponse(taskIds, ranks), nil
	})
	if err != nil {
		switch {
		case errors.Is(err, stores.ErrTaskNotFound),
			errors.Is(err, ErrTaskAlreadyTrashed):
			return nil, err
		default:
			return nil, wrapUnexpectedError(err)
		}
	}

	return response, nil
}

func (service *TaskService) RestoreTask(userId int64, taskId int64) (*transport.RestoreTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("restoring task failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.RestoreTaskResponse, error) {
		task, err := service.taskStore.GetTaskByIdForUser(transaction, taskId, userId)
		if err != nil {
			return nil, err
		}
		if !task.Trashed {
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
		if err := service.taskStore.RestoreTask(transaction, userId, taskId, rank); err != nil {
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
		if task.Trashed {
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
		task.Rank = rank

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

func (service *TaskService) getPreviousTaskRankForTaskDragAndDrop(
	transaction *sqlx.Tx,
	userId int64,
	taskId int64,
	request transport.TaskDragAndDropRequest,
) (string, error) {
	if request.PreviousTaskId == nil {
		return "", nil
	}
	if *request.PreviousTaskId == taskId {
		return "", ErrRequestDataRepresentInvalidDragAndDropGesture
	}

	previousTask, err := service.taskStore.GetTaskByIdForUser(transaction, *request.PreviousTaskId, userId)
	if err != nil {
		if errors.Is(err, stores.ErrTaskNotFound) {
			return "", ErrPreviousTaskNotFound
		}
		return "", err
	}
	previousTaskIsInTargetPartition := previousTask.Trashed == request.Trashed &&
		(request.Trashed || datesAreEqual(previousTask.Date, request.Date) && previousTask.Done == request.Done)
	if !previousTaskIsInTargetPartition {
		return "", ErrRequestDataRepresentInvalidDragAndDropGesture
	}

	return previousTask.Rank, nil
}

func (service *TaskService) isValidDragAndDropAction(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	isMovingFromDatedOpenUntrashedToDatedDoneUntrashed := service.isMovingFromDatedOpenUntrashedToDatedDoneUntrashed(task, request)
	isMovingFromDatedDoneUntrashedToDatedOpenUntrashed := service.isMovingFromDatedDoneUntrashedToDatedOpenUntrashed(task, request)
	isMovingWithinDatedOpenUntrashed := service.isMovingWithinDatedOpenUntrashed(task, request)
	isMovingWithinDatedDoneUntrashed := service.isMovingWithinDatedDoneUntrashed(task, request)

	isValidWorkbenchAction := isMovingFromDatedOpenUntrashedToDatedDoneUntrashed ||
		isMovingFromDatedDoneUntrashedToDatedOpenUntrashed ||
		isMovingWithinDatedOpenUntrashed ||
		isMovingWithinDatedDoneUntrashed

	isMovingFromDatedOpenUntrashedToDatedOpenUntrashed := service.isMovingFromDatedOpenUntrashedToDatedOpenUntrashed(task, request)
	isMovingFromDatedOpenUntrashedToUndatedOpenUntrashed := service.isMovingFromDatedOpenUntrashedToUndatedOpenUntrashed(task, request)
	isMovingFromUndatedOpenUntrashedToDatedOpenUntrashed := service.isMovingFromUndatedOpenUntrashedToDatedOpenUntrashed(task, request)
	isMovingWithinUndatedOpenUntrashed := service.isMovingWithinUndatedOpenUntrashed(task, request)

	isValidBacklogAction := isMovingFromDatedOpenUntrashedToDatedOpenUntrashed ||
		isMovingFromDatedOpenUntrashedToUndatedOpenUntrashed ||
		isMovingFromUndatedOpenUntrashedToDatedOpenUntrashed ||
		isMovingWithinDatedOpenUntrashed ||
		isMovingWithinUndatedOpenUntrashed

	isMovingFromDatedDoneUntrashedToDatedDoneUntrashed := service.isMovingFromDatedDoneUntrashedToDatedDoneUntrashed(task, request)
	isMovingFromDatedDoneUntrashedToDatedDoneTrashed := service.isMovingFromDatedDoneUntrashedToDatedDoneTrashed(task, request)
	isMovingFromTrashedToDatedDoneUntrashed := service.isMovingFromTrashedToDatedDoneUntrashed(task, request)
	isMovingWithinTrashed := service.isMovingWithinTrashed(task, request)

	isValidArchiveAction := isMovingFromDatedDoneUntrashedToDatedDoneUntrashed ||
		isMovingFromDatedDoneUntrashedToDatedDoneTrashed ||
		isMovingFromTrashedToDatedDoneUntrashed ||
		isMovingWithinDatedDoneUntrashed ||
		isMovingWithinTrashed

	return isValidWorkbenchAction || isValidBacklogAction || isValidArchiveAction
}

func (service *TaskService) isMovingFromDatedOpenUntrashedToDatedDoneUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedOpenUntrashed := task.Date != nil && !task.Done && !task.Trashed
	targetIsSameDatedDoneUntrashed := datesAreEqual(task.Date, request.Date) && request.Done && !request.Trashed
	return taskIsDatedOpenUntrashed && targetIsSameDatedDoneUntrashed
}

func (service *TaskService) isMovingFromDatedDoneUntrashedToDatedOpenUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedDoneUntrashed := task.Date != nil && task.Done && !task.Trashed
	targetIsSameDatedOpenUntrashed := datesAreEqual(task.Date, request.Date) && !request.Done && !request.Trashed
	return taskIsDatedDoneUntrashed && targetIsSameDatedOpenUntrashed
}

func (service *TaskService) isMovingWithinDatedOpenUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedOpenUntrashed := task.Date != nil && !task.Done && !task.Trashed
	targetIsSameDatedOpenUntrashed := datesAreEqual(task.Date, request.Date) && !request.Done && !request.Trashed
	return taskIsDatedOpenUntrashed && targetIsSameDatedOpenUntrashed
}

func (service *TaskService) isMovingWithinDatedDoneUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedDoneUntrashed := task.Date != nil && task.Done && !task.Trashed
	targetIsSameDatedDoneUntrashed := datesAreEqual(task.Date, request.Date) && request.Done && !request.Trashed
	return taskIsDatedDoneUntrashed && targetIsSameDatedDoneUntrashed
}

func (service *TaskService) isMovingFromDatedOpenUntrashedToDatedOpenUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedOpenUntrashed := task.Date != nil && !task.Done && !task.Trashed
	targetIsDifferentDatedOpenUntrashed := request.Date != nil && !datesAreEqual(task.Date, request.Date) && !request.Done && !request.Trashed
	return taskIsDatedOpenUntrashed && targetIsDifferentDatedOpenUntrashed
}

func (service *TaskService) isMovingFromDatedOpenUntrashedToUndatedOpenUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedOpenUntrashed := task.Date != nil && !task.Done && !task.Trashed
	targetIsUndatedOpenUntrashed := request.Date == nil && !request.Done && !request.Trashed
	return taskIsDatedOpenUntrashed && targetIsUndatedOpenUntrashed
}

func (service *TaskService) isMovingFromUndatedOpenUntrashedToDatedOpenUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsUndatedOpenUntrashed := task.Date == nil && !task.Done && !task.Trashed
	targetIsDatedOpenUntrashed := request.Date != nil && !request.Done && !request.Trashed
	return taskIsUndatedOpenUntrashed && targetIsDatedOpenUntrashed
}

func (service *TaskService) isMovingWithinUndatedOpenUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsUndatedOpenUntrashed := task.Date == nil && !task.Done && !task.Trashed
	targetIsUndatedOpenUntrashed := request.Date == nil && !request.Done && !request.Trashed
	return taskIsUndatedOpenUntrashed && targetIsUndatedOpenUntrashed
}

func (service *TaskService) isMovingFromDatedDoneUntrashedToDatedDoneUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedDoneUntrashed := task.Date != nil && task.Done && !task.Trashed
	targetIsDifferentDatedDoneUntrashed := request.Date != nil && !datesAreEqual(task.Date, request.Date) && request.Done && !request.Trashed
	return taskIsDatedDoneUntrashed && targetIsDifferentDatedDoneUntrashed
}

func (service *TaskService) isMovingFromDatedDoneUntrashedToDatedDoneTrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedDoneUntrashed := task.Date != nil && task.Done && !task.Trashed
	targetIsDatedDoneTrashed := request.Date != nil && request.Done && request.Trashed
	return taskIsDatedDoneUntrashed && targetIsDatedDoneTrashed
}

func (service *TaskService) isMovingFromTrashedToDatedDoneUntrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	taskIsDatedDoneTrashed := task.Date != nil && task.Done && task.Trashed
	taskIsDatedOpenTrashed := task.Date != nil && !task.Done && task.Trashed
	taskIsUndatedOpenTrashed := task.Date == nil && !task.Done && task.Trashed
	targetIsDatedDoneUntrashed := request.Date != nil && request.Done && !request.Trashed
	return (taskIsDatedDoneTrashed || taskIsDatedOpenTrashed || taskIsUndatedOpenTrashed) && targetIsDatedDoneUntrashed
}

func (service *TaskService) isMovingWithinTrashed(task *domain.Task, request transport.TaskDragAndDropRequest) bool {
	isMovingWithinDatedDoneTrashed := task.Date != nil && task.Done && task.Trashed &&
		datesAreEqual(task.Date, request.Date) && request.Done && request.Trashed
	isMovingWithinDatedOpenTrashed := task.Date != nil && !task.Done && task.Trashed &&
		datesAreEqual(task.Date, request.Date) && !request.Done && request.Trashed
	isMovingWithinUndatedOpenTrashed := task.Date == nil && !task.Done && task.Trashed &&
		datesAreEqual(task.Date, request.Date) && !request.Done && request.Trashed
	return isMovingWithinDatedDoneTrashed || isMovingWithinDatedOpenTrashed || isMovingWithinUndatedOpenTrashed
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
		if task.Trashed {
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
			task.Rank = rank
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
			Rank: task.Rank,
		})
	}
	return response
}

func (service *TaskService) createBatchOpenTasksResponse(tasks []*domain.Task) []transport.BatchOpenTaskResponse {
	response := make([]transport.BatchOpenTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, transport.BatchOpenTaskResponse{
			Id:   task.Id,
			Rank: task.Rank,
		})
	}
	return response
}

func (service *TaskService) createBatchTrashTasksResponse(taskIds []int64, ranks []string) []transport.BatchTrashTaskResponse {
	response := make([]transport.BatchTrashTaskResponse, 0, len(taskIds))
	for index, taskId := range taskIds {
		response = append(response, transport.BatchTrashTaskResponse{
			Id:   taskId,
			Rank: ranks[index],
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
