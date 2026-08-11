package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksToggleDoneUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/toggle-done?"+url.Values{"rank": {"a0"}}.Encode(),
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

func TestTasksToggleDoneInvalidId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/invalid/toggle-done?"+url.Values{"rank": {"a0"}}.Encode(),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_task_id",
		"Task id must be a positive integer.",
	)
}

func TestTasksToggleDoneInvalidRank(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/toggle-done?"+url.Values{"rank": {"!"}}.Encode(),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter",
		"URL contains an invalid query parameter",
	)
}

func TestTasksToggleDoneMissingRank(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	task := tasks[0]

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"missing_query_parameter",
		"URL is missing a required query parameter.",
	)
}

func TestTasksToggleDoneTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/toggle-done?"+url.Values{"rank": {"a0"}}.Encode(),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"task_not_found",
		"No task exists for the provided id.",
	)
}

func TestTasksToggleDoneRankAlreadyExists(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, cookies...)
	openTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := openTasks[0]
	require.NotNil(t, doneTasks[0].Rank, "expected done task to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/toggle-done?%s",
			task.Id,
			url.Values{"rank": {*doneTasks[0].Rank}}.Encode(),
		),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusConflict,
		"task_rank_already_exists",
		"The provided rank is already in use.",
	)
}

func TestTasksToggleDoneForDatedOpenUntrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := tasks[0]
	require.NotNil(t, task.Rank, "expected task to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/toggle-done?%s",
			task.Id,
			url.Values{"rank": {*task.Rank}}.Encode(),
		),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}

func TestTasksToggleDoneForDatedDoneUntrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := tasks[0]
	require.NotNil(t, task.Rank, "expected task to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/toggle-done?%s",
			task.Id,
			url.Values{"rank": {*task.Rank}}.Encode(),
		),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}

func TestTasksToggleDoneForUndatedOpenUntrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	task := tasks[0]
	require.NotNil(t, task.Rank, "expected task to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/toggle-done?%s",
			task.Id,
			url.Values{"rank": {*task.Rank}}.Encode(),
		),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}

func TestTasksToggleDoneForUndatedOpenTrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range tasks {
		if !candidate.Done && candidate.Date == nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open undated trashed task")
	require.Nil(t, task.Rank, "expected trashed task not to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}

func TestTasksToggleDoneForDatedDoneTrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range tasks {
		if candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected a done dated trashed task")
	require.Nil(t, task.Rank, "expected trashed task not to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}

func TestTasksToggleDoneForDatedOpenTrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range tasks {
		if !candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open dated trashed task")
	require.Nil(t, task.Rank, "expected trashed task not to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}
