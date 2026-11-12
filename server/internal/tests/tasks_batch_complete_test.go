package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksBatchCompleteUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{1}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksBatchCompleteEmptyTaskIds(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at least 1 items.")
}

func TestTasksBatchCompleteTooManyTaskIds(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	taskIds := make([]int64, 51)
	for index := range taskIds {
		taskIds[index] = int64(index + 1)
	}
	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: taskIds})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at most 50 items.")
}

func TestTasksBatchCompleteDuplicateTaskId(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{1, 1}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must not contain duplicate values.")
}

func TestTasksBatchCompleteInvalidTaskId(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{0}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds[0]", "taskIds[0] must be greater than 0.")
}

func TestTasksBatchCompleteOneTaskOwnedByDifferentUser(t *testing.T) {
	sut := support.PrepareTestEnvironment(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	app := sut.App
	firstSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	firstUserCookies := firstSignUpResp.Result().Cookies()
	secondSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser2Password)
	secondUserCookies := secondSignUpResp.Result().Cookies()

	firstCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", firstUserCookies...)
	require.Equal(t, http.StatusCreated, firstCategoryResp.Code)
	var firstCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(firstCategoryResp.Body).Decode(&firstCategory), "failed to decode first category response")

	secondCategoryResp := support.MakeCreateCategoryRequest(t, app, "Personal", "#3b82f6", secondUserCookies...)
	require.Equal(t, http.StatusCreated, secondCategoryResp.Code)
	var secondCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(secondCategoryResp.Body).Decode(&secondCategory), "failed to decode second category response")

	firstUserTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, firstCategory.Id, firstUserCookies...)
	secondUserTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, secondCategory.Id, secondUserCookies...)
	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{firstUserTasks[0].Id, secondUserTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, firstUserCookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"one_of_provided_tasks_not_found",
		"One of the provided tasks could not be found.",
	)
}

func TestTasksBatchCompleteOneTaskAlreadyDone(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	openTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	completedTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{openTasks[0].Id, completedTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"one_of_provided_tasks_already_completed",
		"One of the provided tasks is already completed.",
	)
}

func TestTasksBatchCompleteOneTaskTrashed(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	var openTrashedTask tasksTransport.TrashedTaskResponse
	for _, task := range trashedTasks {
		if !task.Done {
			openTrashedTask = task
			break
		}
	}
	require.NotZero(t, openTrashedTask.Id, "expected an open trashed task")

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{openTrashedTask.Id}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"one_of_provided_tasks_is_trashed",
		"Trashed tasks can not be part of a batch operation.",
	)
}

func TestTasksBatchCompleteMultipleTasks(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	undatedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{undatedTasks[0].Id, datedTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var respBody []tasksTransport.BatchCompleteTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")
	require.Len(t, respBody, 2)

	assert.ElementsMatch(t,
		[]int64{undatedTasks[0].Id, datedTasks[0].Id},
		[]int64{respBody[0].Id, respBody[1].Id},
	)
	for _, task := range respBody {
		assert.False(t, task.Date.IsZero(), "expected completed task to have a date")
		assert.NotEmpty(t, task.Rank, "expected untrashed task to have a rank")
	}
}
