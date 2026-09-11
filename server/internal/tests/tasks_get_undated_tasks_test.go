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

func TestTasksGetUndatedTasksUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/undated", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetUndatedTasksInvalidRank(t *testing.T) {
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
		"/api/tasks/undated?"+url.Values{"rank": {"!"}}.Encode(),
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

func TestTasksGetUndatedTasksNoCursor(t *testing.T) {
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

	expectedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/undated", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetUndatedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, toUndatedTaskResponses(t, expectedTasks), respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "expected no additional tasks")
}

func TestTasksGetUndatedTasksCursor(t *testing.T) {
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

	tasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	cursor := tasks[0]
	require.NotNil(t, cursor.Rank, "expected cursor task to have a rank")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/undated?"+url.Values{"rank": {*cursor.Rank}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetUndatedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, toUndatedTaskResponses(t, tasks[1:]), respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "expected no additional tasks")
}

func TestTasksGetUndatedTasksHasMore(t *testing.T) {
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
	expectedTasks := make([]tasksTransport.UndatedTaskResponse, 0, taskBatchSize)
	for index := range taskBatchSize + 1 {
		title := fmt.Sprintf("Task %d", index+1)
		createTaskResp := support.MakeCreateTaskRequest(t, app, title, nil, nil, category.Id, cookies...)
		require.Equal(t, http.StatusCreated, createTaskResp.Code)
		var task tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(createTaskResp.Body).Decode(&task), "failed to decode task response")
		if index < taskBatchSize {
			expectedTasks = append(expectedTasks, tasksTransport.UndatedTaskResponse{
				Id:         task.Id,
				CategoryId: category.Id,
				Title:      title,
				Rank:       task.Rank,
			})
		}
	}

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/undated", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetUndatedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.True(t, respBody.HasMore, "expected additional tasks")
}

func toUndatedTaskResponses(t *testing.T, tasks []tasksTransport.TaskResponse) []tasksTransport.UndatedTaskResponse {
	t.Helper()

	responses := make([]tasksTransport.UndatedTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		require.NotNil(t, task.Rank, "expected undated task to have a rank")
		responses = append(responses, tasksTransport.UndatedTaskResponse{
			Id:          task.Id,
			CategoryId:  task.CategoryId,
			Title:       task.Title,
			Description: task.Description,
			Rank:        *task.Rank,
		})
	}

	return responses
}
