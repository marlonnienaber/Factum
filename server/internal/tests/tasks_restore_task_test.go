package tests

import (
	"encoding/json"
	"fmt"
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

func TestTasksRestoreTaskUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/restore?"+url.Values{"rank": {"a0"}}.Encode(),
		nil,
	)

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
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/invalid/restore?rank=a0",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_task_id",
		"Task id must be a positive integer.",
	)
}

func TestTasksRestoreTaskNoRank(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/restore",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"missing_query_parameter",
		"URL is missing a required query parameter.",
	)
}

func TestTasksRestoreTaskInvalidRank(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/restore?"+url.Values{"rank": {"!"}}.Encode(),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter",
		"URL contains an invalid query parameter",
	)
}

func TestTasksRestoreTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/restore?"+url.Values{"rank": {"a0"}}.Encode(),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"task_not_found",
		"No task exists for the provided id.",
	)
}

func TestTasksRestoreTaskAlreadyRestored(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	createTaskResp := support.MakeCreateTaskRequest(
		t,
		app,
		"Task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createTaskResp.Code)

	var task tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(createTaskResp.Body).Decode(&task),
		"failed to decode task response",
	)
	require.NotNil(t, task.Rank, "expected task rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/restore?%s",
			task.Id,
			url.Values{"rank": {*task.Rank}}.Encode(),
		),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"task_already_restored",
		"Task is already restored.",
	)
}

func TestTasksRestoreTaskRankAlreadyExists(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	createFirstTaskResp := support.MakeCreateTaskRequest(
		t,
		app,
		"First task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createFirstTaskResp.Code)

	var firstTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(createFirstTaskResp.Body).Decode(&firstTask),
		"failed to decode first task response",
	)
	require.NotNil(t, firstTask.Rank, "expected first task rank")

	createSecondTaskResp := support.MakeCreateTaskRequest(
		t,
		app,
		"Second task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createSecondTaskResp.Code)

	var secondTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(createSecondTaskResp.Body).Decode(&secondTask),
		"failed to decode second task response",
	)

	trashResp := support.MakeTrashTaskRequest(
		t,
		app,
		secondTask.Id,
		time.Now().UTC(),
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusNoContent, trashResp.Code)

	restoreResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/restore?%s",
			secondTask.Id,
			url.Values{"rank": {*firstTask.Rank}}.Encode(),
		),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		restoreResp,
		http.StatusConflict,
		"task_rank_already_exists",
		"The provided rank is already in use.",
	)
}

func TestTasksRestoreTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	createTaskResp := support.MakeCreateTaskRequest(
		t,
		app,
		"Task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createTaskResp.Code)

	var task tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(createTaskResp.Body).Decode(&task),
		"failed to decode task response",
	)
	require.NotNil(t, task.Rank, "expected task rank")

	trashResp := support.MakeTrashTaskRequest(
		t,
		app,
		task.Id,
		time.Now().UTC(),
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusNoContent, trashResp.Code)

	restoreResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/restore?%s",
			task.Id,
			url.Values{"rank": {*task.Rank}}.Encode(),
		),
		nil,
		signUpResp.Result().Cookies()...,
	)

	require.Equal(t, http.StatusNoContent, restoreResp.Code)
	assert.Empty(t, restoreResp.Body.String(), "expected an empty response body")
}
