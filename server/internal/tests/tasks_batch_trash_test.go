package tests

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var batchTrashEncodedTrashedDate = url.Values{
	"trashedDate": {time.Now().UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)},
}.Encode()

func TestTasksBatchTrashUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{TaskIds: []int64{1}})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksBatchTrashEmptyTaskIds(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{TaskIds: []int64{}})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at least 1 items.")
}

func TestTasksBatchTrashTooManyTaskIds(t *testing.T) {
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
	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{TaskIds: taskIds})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must contain at most 50 items.")
}

func TestTasksBatchTrashDuplicateTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{TaskIds: []int64{1, 1}})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds", "taskIds must not contain duplicate values.")
}

func TestTasksBatchTrashInvalidTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{TaskIds: []int64{0}})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "taskIds[0]", "taskIds[0] must be greater than 0.")
}

func TestTasksBatchTrashOneTaskAlreadyTrashed(t *testing.T) {
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
	untrashedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)

	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{
		TaskIds: []int64{trashedTasks[0].Id, untrashedTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"one_of_provided_tasks_already_trashed",
		"One of the provided tasks is already trashed.",
	)
}

func TestTasksBatchTrashOneTaskOwnedByDifferentUser(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{
		TaskIds: []int64{firstUserTasks[0].Id, secondUserTasks[0].Id},
	})
	require.NoError(t, err, "serializing request failed")
	path := "/api/tasks/batch-trash?" + batchTrashEncodedTrashedDate
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, firstUserCookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"one_of_provided_tasks_not_found",
		"One of the provided tasks could not be found.",
	)
}

func TestTasksBatchTrashMultipleTasks(t *testing.T) {
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

	undatedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedOpenTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	datedDoneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	taskIds := []int64{undatedTasks[0].Id, datedOpenTasks[0].Id, datedDoneTasks[0].Id}
	payload, err := json.Marshal(tasksTransport.BatchTrashTasksRequest{TaskIds: taskIds})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/batch-trash", payload, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var respBody []tasksTransport.BatchTrashTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")
	require.Len(t, respBody, len(taskIds))

	for index, task := range respBody {
		assert.Equal(t, taskIds[index], task.Id, "unexpected task id")
		assert.NotEmpty(t, task.Rank, "expected trashed task to have a rank")
		if index > 0 {
			assert.Less(t, respBody[index-1].Rank, task.Rank, "expected tasks to retain client order")
		}
	}
}
