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
	ErrTaskAlreadyTrashed                            = errors.New("task is already trashed")
	ErrTaskAlreadyRestored                           = errors.New("task is already restored")
	ErrTaskAlreadyCompleted                          = errors.New("task is already completed")
	ErrTaskAlreadyOpen                               = errors.New("task is already open")
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

func (service *TaskService) CreateTask(userId int64, request transport.CreateTaskRequest) (*transport.CreateTaskResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("creating task failed")

	task, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.CreateTaskResponse, error) {
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

func (service *TaskService) BatchCompleteTasks(userId int64, taskIds []int64) (*transport.BatchCompleteResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch completing tasks failed")
	currentDate := shared.NewDate(time.Now())

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.BatchCompleteResponse, error) {
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

		return service.createBatchCompleteTasksResponse(tasks), nil
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

func (service *TaskService) BatchOpenTasks(userId int64, taskIds []int64) (*transport.BatchOpenResponse, error) {
	wrapUnexpectedError := shared.GetWrapUnexpectedErrorHelper("batch opening tasks failed")

	response, err := shared.ExecuteAsSerializableTransaction(service.db, func(transaction *sqlx.Tx) (*transport.BatchOpenResponse, error) {
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

		return service.createBatchOpenTasksResponse(tasks), nil
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

func (service *TaskService) createBatchCompleteTasksResponse(tasks []domain.Task) *transport.BatchCompleteResponse {
	response := &transport.BatchCompleteResponse{
		DatedUntrashedDoneTasks: make([]transport.TaskResponse, 0),
		TrashedTasks:            make([]transport.TaskResponse, 0),
	}
	for index := range tasks {
		task := &tasks[index]
		taskResponse := *task.ToTaskResponse()
		switch {
		case task.TrashedDate != nil:
			response.TrashedTasks = append(response.TrashedTasks, taskResponse)
		default:
			response.DatedUntrashedDoneTasks = append(response.DatedUntrashedDoneTasks, taskResponse)
		}
	}
	return response
}

func (service *TaskService) createBatchOpenTasksResponse(tasks []domain.Task) *transport.BatchOpenResponse {
	response := &transport.BatchOpenResponse{
		DatedUntrashedOpenTasks: make([]transport.TaskResponse, 0),
		TrashedTasks:            make([]transport.TaskResponse, 0),
	}
	for index := range tasks {
		task := &tasks[index]
		taskResponse := *task.ToTaskResponse()
		switch {
		case task.TrashedDate != nil:
			response.TrashedTasks = append(response.TrashedTasks, taskResponse)
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
