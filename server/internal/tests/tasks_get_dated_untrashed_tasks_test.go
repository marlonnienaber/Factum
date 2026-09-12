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

func TestTasksGetDatedUntrashedTasksUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{"done": {"true"}}.Encode(),
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

func TestTasksGetDatedUntrashedTasksNoDoneProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed", nil, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"missing_query_parameter",
		"URL is missing a required query parameter.",
	)
}

func TestTasksGetDatedUntrashedTasksInvalidDoneProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{"done": {"invalid"}}.Encode(),
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

func TestTasksGetDatedUntrashedTasksInvalidDateProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{
			"done": {"true"},
			"date": {"invalid"},
		}.Encode(),
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

func TestTasksGetDatedUntrashedTasksInvalidRankProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{
			"done": {"true"},
			"rank": {"!"},
		}.Encode(),
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

func TestTasksGetDatedUntrashedTasksIncompleteCursor(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{
			"done": {"true"},
			"date": {"2026-03-25"},
		}.Encode(),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter_combination",
		"URL contains an invalid query parameter combination.",
	)
}

func TestTasksGetDatedUntrashedTasksDoneWithoutCursor(t *testing.T) {
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

	expectedTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{"done": {"true"}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedTasksDoneWithCursor(t *testing.T) {
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

	tasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	cursor := tasks[0]
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{
			"done": {"true"},
			"date": {cursor.Date.Format(time.DateOnly)},
			"rank": {cursor.Rank},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, tasks[1:], respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedTasksOpenWithoutCursor(t *testing.T) {
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

	expectedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{"done": {"false"}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedTasksOpenWithCursor(t *testing.T) {
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

	tasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	cursor := tasks[0]
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{
			"done": {"false"},
			"date": {cursor.Date.Format(time.DateOnly)},
			"rank": {cursor.Rank},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, tasks[1:], respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedOpenTasksHasMore(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	const taskBatchSize = 50
	expectedTasks := make([]tasksTransport.DatedUntrashedTaskResponse, 0, taskBatchSize)
	for index := range taskBatchSize + 1 {
		title := fmt.Sprintf("Task %d", index+1)
		createTaskResp := support.MakeCreateTaskRequest(t, app, title, nil, &support.TestTaskDate, category.Id, cookies...)
		require.Equal(t, http.StatusCreated, createTaskResp.Code)
		var task tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(createTaskResp.Body).Decode(&task), "failed to decode task response")
		if index < taskBatchSize {
			expectedTasks = append(expectedTasks, tasksTransport.DatedUntrashedTaskResponse{
				Id:         task.Id,
				CategoryId: category.Id,
				Title:      title,
				Date:       support.TestTaskDate,
				Rank:       task.Rank,
			})
		}
	}

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{"done": {"false"}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.True(t, respBody.HasMore, "expected additional tasks")
}

func TestTasksGetDatedUntrashedDoneTasksHasMore(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	const taskBatchSize = 50
	expectedTasks := make([]tasksTransport.DatedUntrashedTaskResponse, 0, taskBatchSize)
	for index := range taskBatchSize + 1 {
		title := fmt.Sprintf("Task %d", index+1)
		createTaskResp := support.MakeCreateTaskRequest(t, app, title, nil, &support.TestTaskDate, category.Id, cookies...)
		require.Equal(t, http.StatusCreated, createTaskResp.Code)
		var task tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(createTaskResp.Body).Decode(&task), "failed to decode task response")

		toggleDoneResp := support.MakeToggleDoneRequest(t, app, task.Id, cookies...)
		require.Equal(t, http.StatusOK, toggleDoneResp.Code)
		var toggleDoneResponse tasksTransport.ToggleTaskDoneResponse
		require.NoError(t, json.NewDecoder(toggleDoneResp.Body).Decode(&toggleDoneResponse), "failed to decode toggle task done response")
		require.NotNil(t, toggleDoneResponse.Rank, "expected toggled task to have a rank")
		if index < taskBatchSize {
			expectedTasks = append(expectedTasks, tasksTransport.DatedUntrashedTaskResponse{
				Id:         task.Id,
				CategoryId: category.Id,
				Title:      title,
				Date:       toggleDoneResponse.Date,
				Rank:       *toggleDoneResponse.Rank,
				Done:       true,
			})
		}
	}

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed?"+url.Values{"done": {"true"}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.True(t, respBody.HasMore, "expected additional tasks")
}
