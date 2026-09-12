package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/require"
	"roci.dev/fracdex"
)

func TestTasksRestoreTaskUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/restore", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksRestoreTaskInvalidId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/invalid/restore", nil, signUpResp.Result().Cookies()...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_task_id", "Task id must be a positive integer.")
}

func TestTasksRestoreTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/restore", nil, signUpResp.Result().Cookies()...)

	support.AssertErrorResponse(t, resp, http.StatusNotFound, "task_not_found", "No task exists for the provided id.")
}

func TestTasksRestoreTaskAlreadyRestored(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	tasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/restore", tasks[0].Id), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "task_already_restored", "Task is already restored.")
}

func TestTasksRestoreOpenUndatedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	destinationTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TrashedTaskResponse
	for _, candidate := range trashedTasks {
		if candidate.Date == nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open undated trashed task")
	lastDestinationTask := destinationTasks[len(destinationTasks)-1]
	expectedRank, err := fracdex.KeyBetween(lastDestinationTask.Rank, "")
	require.NoError(t, err, "generating expected rank failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/restore", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.RestoreTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode restore task response")
	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksRestoreOpenDatedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	destinationTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TrashedTaskResponse
	for _, candidate := range trashedTasks {
		if candidate.Date != nil && !candidate.Done {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open dated trashed task")
	lastDestinationTask := destinationTasks[len(destinationTasks)-1]
	expectedRank, err := fracdex.KeyBetween(lastDestinationTask.Rank, "")
	require.NoError(t, err, "generating expected rank failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/restore", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.RestoreTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode restore task response")
	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksRestoreDoneDatedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	destinationTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TrashedTaskResponse
	for _, candidate := range trashedTasks {
		if candidate.Done {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected a done dated trashed task")
	lastDestinationTask := destinationTasks[len(destinationTasks)-1]
	expectedRank, err := fracdex.KeyBetween(lastDestinationTask.Rank, "")
	require.NoError(t, err, "generating expected rank failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/restore", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.RestoreTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode restore task response")
	require.Equal(t, expectedRank, response.Rank)
}
