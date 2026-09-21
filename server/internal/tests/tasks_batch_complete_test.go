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

func TestTasksBatchCompleteUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at least 1 items.")
}

func TestTasksBatchCompleteTooManyTaskIds(t *testing.T) {
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
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at most 50 items.")
}

func TestTasksBatchCompleteDuplicateTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{1, 1}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must not contain duplicate values.")
}

func TestTasksBatchCompleteInvalidTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{TaskIds: []int64{0}})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds[0]", "taskIds[0] must be greater than 0.")
}

func TestTasksBatchCompleteOneTaskOwnedByDifferentUser(t *testing.T) {
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

func TestTasksBatchCompleteTasksFromAllValidPartitions(t *testing.T) {
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

	undatedUntrashedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedUntrashedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)

	var undatedTrashedTask tasksTransport.TrashedTaskResponse
	var datedTrashedTask tasksTransport.TrashedTaskResponse
	for _, task := range trashedTasks {
		if task.Done {
			continue
		}
		if task.Date == nil {
			undatedTrashedTask = task
		} else {
			datedTrashedTask = task
		}
	}
	require.NotZero(t, undatedTrashedTask.Id, "expected an open undated trashed task")
	require.NotZero(t, datedTrashedTask.Id, "expected an open dated trashed task")

	payload, err := json.Marshal(tasksTransport.BatchUpdateDoneRequest{
		TaskIds: []int64{
			undatedUntrashedTasks[0].Id,
			datedUntrashedTasks[0].Id,
			undatedTrashedTask.Id,
			datedTrashedTask.Id,
		},
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-complete", payload, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.BatchCompleteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")
	require.Len(t, respBody.DatedUntrashedDoneTasks, 2)
	require.Len(t, respBody.TrashedTasks, 2)

	assert.ElementsMatch(t,
		[]int64{undatedUntrashedTasks[0].Id, datedUntrashedTasks[0].Id},
		[]int64{respBody.DatedUntrashedDoneTasks[0].Id, respBody.DatedUntrashedDoneTasks[1].Id},
	)
	for _, task := range respBody.DatedUntrashedDoneTasks {
		assert.False(t, task.Date.IsZero(), "expected completed task to have a date")
		assert.NotEmpty(t, task.Rank, "expected untrashed task to have a rank")
	}

	assert.ElementsMatch(t,
		[]int64{undatedTrashedTask.Id, datedTrashedTask.Id},
		[]int64{respBody.TrashedTasks[0].Id, respBody.TrashedTasks[1].Id},
	)
	for _, task := range respBody.TrashedTasks {
		assert.True(t, task.Done, "expected task to be completed")
		assert.NotNil(t, task.Date, "expected completed task to have a date")
		assert.False(t, task.TrashedDate.IsZero(), "expected task to remain trashed")
	}
}
