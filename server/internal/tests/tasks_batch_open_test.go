package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksBatchOpenUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{1}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksBatchOpenEmptyTaskIds(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at least 1 items.")
}

func TestTasksBatchOpenTooManyTaskIds(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	taskIds := make([]int64, 51)
	for index := range taskIds {
		taskIds[index] = int64(index + 1)
	}
	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: taskIds})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at most 50 items.")
}

func TestTasksBatchOpenDuplicateTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{1, 1}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must not contain duplicate values.")
}

func TestTasksBatchOpenInvalidTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{0}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds[0]", "taskIds[0] must be greater than 0.")
}

func TestTasksBatchOpenOneTaskOwnedByDifferentUser(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	firstSignUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, firstSignUpResp.Code)
	firstUserCookies := firstSignUpResp.Result().Cookies()
	secondSignUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser2Email,
		Password: support.TestUser2Password,
	})
	require.Equal(t, http.StatusCreated, secondSignUpResp.Code)
	secondUserCookies := secondSignUpResp.Result().Cookies()

	firstCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", firstUserCookies...)
	require.Equal(t, http.StatusCreated, firstCategoryResp.Code)
	var firstCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(firstCategoryResp.Body).Decode(&firstCategory), "failed to decode first category response")

	secondCategoryResp := support.MakeCreateCategoryRequest(t, app, "Personal", "#3b82f6", secondUserCookies...)
	require.Equal(t, http.StatusCreated, secondCategoryResp.Code)
	var secondCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(secondCategoryResp.Body).Decode(&secondCategory), "failed to decode second category response")

	firstUserTasks := support.PrepareDoneDatedUntrashedTasks(t, app, firstCategory.Id, support.TestTaskDate, firstUserCookies...)
	secondUserTasks := support.PrepareDoneDatedUntrashedTasks(t, app, secondCategory.Id, support.TestTaskDate, secondUserCookies...)

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{firstUserTasks[0].Id, secondUserTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, firstUserCookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"one_of_provided_tasks_not_found",
		"One of the provided tasks could not be found.",
	)
}

func TestTasksBatchOpenOneTaskAlreadyOpen(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	completedTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	openTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{completedTasks[0].Id, openTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"one_of_provided_tasks_already_open",
		"One of the provided tasks is already open.",
	)
}

func TestTasksBatchOpenOneTaskTrashed(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	completedUntrashedTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	var completedTrashedTask tasksTransport.TrashedTaskResponse
	for _, task := range trashedTasks {
		if task.Done {
			completedTrashedTask = task
			break
		}
	}
	require.NotZero(t, completedTrashedTask.Id, "expected a completed trashed task")
	require.NotNil(t, completedTrashedTask.Date, "expected completed task to have a date")

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{completedUntrashedTasks[0].Id, completedTrashedTask.Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"one_of_provided_tasks_is_trashed",
		"Trashed tasks can not be part of a batch operation.",
	)
}

func TestTasksBatchOpenMultipleTasks(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	datedDoneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{datedDoneTasks[0].Id, datedDoneTasks[1].Id},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-open", payload, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var respBody []tasksTransport.BatchOpenTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")
	require.Len(t, respBody, 2)

	assert.ElementsMatch(t,
		[]int64{datedDoneTasks[0].Id, datedDoneTasks[1].Id},
		[]int64{respBody[0].Id, respBody[1].Id},
	)
	for _, task := range respBody {
		assert.NotEmpty(t, task.Rank, "expected opened task to have a rank")
	}
}
