package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksUpdateTaskUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         1,
		Title:      "Task",
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksUpdateTaskNoId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(map[string]any{
		"title":      "Task",
		"categoryId": 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "id", "id must be greater than 0.")
}

func TestTasksUpdateTaskInvalidId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         -1,
		Title:      "Task",
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "id", "id must be greater than 0.")
}

func TestTasksUpdateTaskNoTitle(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(map[string]any{
		"id":         1,
		"categoryId": 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "title", "title must be at least 1 characters long.")
}

func TestTasksUpdateTaskTitleTooShort(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         1,
		Title:      "",
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "title", "title must be at least 1 characters long.")
}

func TestTasksUpdateTaskTitleTooLong(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         1,
		Title:      strings.Repeat("a", maximumTaskTitleLength+1),
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "title", "title must be at most 100 characters long.")
}

func TestTasksUpdateTaskDescriptionTooShort(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	description := ""
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:          1,
		Title:       "Task",
		Description: &description,
		CategoryId:  1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"description",
		"description must be at least 1 characters long.",
	)
}

func TestTasksUpdateTaskDescriptionTooLong(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	description := strings.Repeat("a", maximumTaskDescriptionLength+1)
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:          1,
		Title:       "Task",
		Description: &description,
		CategoryId:  1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"description",
		"description must be at most 2000 characters long.",
	)
}

func TestTasksUpdateTaskNoCategoryId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(map[string]any{
		"id":    1,
		"title": "Task",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"categoryId",
		"categoryId must be greater than 0.",
	)
}

func TestTasksUpdateTaskInvalidCategoryId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         1,
		Title:      "Task",
		CategoryId: -1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"categoryId",
		"categoryId must be greater than 0.",
	)
}

func TestTasksUpdateTaskTaskNotFound(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         1,
		Title:      "Task",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"task_not_found",
		"No task exists for the provided id.",
	)
}

func TestTasksUpdateTaskCategoryNotFound(t *testing.T) {
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

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: category.Id + 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"category_not_found",
		"No category exists for the provided categoryId.",
	)
}

func TestTasksUpdateTaskChangeTitle(t *testing.T) {
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

	taskResp := support.MakeCreateTaskRequest(t, app, "Old title", nil, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "New title",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, "New title", respBody.Title, "unexpected task title")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskAddDescription(t *testing.T) {
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

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	description := "Description"
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:          task.Id,
		Title:       task.Title,
		Description: &description,
		CategoryId:  category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &description, respBody.Description, "unexpected task description")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskRemoveDescription(t *testing.T) {
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

	description := "Description"
	taskResp := support.MakeCreateTaskRequest(t, app, "Task", &description, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Description, "expected task description to be removed")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskChangeDescription(t *testing.T) {
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

	oldDescription := "Old description"
	taskResp := support.MakeCreateTaskRequest(t, app, "Task", &oldDescription, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	newDescription := "New description"
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:          task.Id,
		Title:       task.Title,
		Description: &newDescription,
		CategoryId:  category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &newDescription, respBody.Description, "unexpected task description")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskFromOpenWithoutDateUntrashedToDate(t *testing.T) {
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

	undatedTasks := prepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedTasks := prepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := undatedTasks[0]
	require.NotNil(t, datedTasks[0].Date, "expected dated task date")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		Date:       datedTasks[0].Date,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, datedTasks[0].Date, respBody.Date, "unexpected task date")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromOpenWithDateUntrashedToNoDate(t *testing.T) {
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

	prepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedTasks := prepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := datedTasks[0]

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromDoneWithDateUntrashedToNoDate(t *testing.T) {
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

	prepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	doneTasks := prepareDoneDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := doneTasks[0]

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to become open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromDoneWithDateTrashedToNoDateTrashed(t *testing.T) {
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

	trashedTasks := prepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range trashedTasks {
		if candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected a done dated trashed task")
	require.Nil(t, task.Rank, "expected trashed task not to have a rank")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to become open")
	assert.Nil(t, respBody.Rank, "expected trashed task not to have a rank")
}

func TestTasksUpdateTaskFromOpenWithDateTrashedToNoDateTrashed(t *testing.T) {
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

	trashedTasks := prepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range trashedTasks {
		if !candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open dated trashed task")
	require.Nil(t, task.Rank, "expected trashed task not to have a rank")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.Nil(t, respBody.Rank, "expected trashed task not to have a rank")
}

func TestTasksUpdateTaskFromOpenWithoutDateTrashedToDateTrashed(t *testing.T) {
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

	trashedTasks := prepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range trashedTasks {
		if !candidate.Done && candidate.Date == nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open undated trashed task")
	require.Nil(t, task.Rank, "expected trashed task not to have a rank")
	targetDate, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		Date:       &targetDate,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.Nil(t, respBody.Rank, "expected trashed task not to have a rank")
}

func TestTasksUpdateTaskFromOpenWithDateToOtherDate(t *testing.T) {
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

	datedTasks := prepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := datedTasks[0]
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		Date:       &targetDate,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromDoneWithDateToOtherDate(t *testing.T) {
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

	doneTasks := prepareDoneDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := doneTasks[0]
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		Date:       &targetDate,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.True(t, respBody.Done, "expected task to remain done")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskChangeCategory(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	oldCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", cookies...)
	require.Equal(t, http.StatusCreated, oldCategoryResp.Code)
	var oldCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(oldCategoryResp.Body).Decode(&oldCategory), "failed to decode category response")

	newCategoryResp := support.MakeCreateCategoryRequest(t, app, "Personal", cookies...)
	require.Equal(t, http.StatusCreated, newCategoryResp.Code)
	var newCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(newCategoryResp.Body).Decode(&newCategory), "failed to decode category response")

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, oldCategory.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      task.Title,
		CategoryId: newCategory.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, newCategory.Id, respBody.CategoryId, "unexpected category id")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

const preparedTasksCount = 3

func prepareOpenUndatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	tasks := make([]tasksTransport.TaskResponse, 0, preparedTasksCount)
	for range preparedTasksCount {
		resp := support.MakeCreateTaskRequest(t, app, "Undated task", nil, nil, categoryId, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)

		var task tasksTransport.TaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&task), "failed to decode task response")
		tasks = append(tasks, task)
	}
	return tasks
}

func prepareOpenDatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	date, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")

	tasks := make([]tasksTransport.TaskResponse, 0, preparedTasksCount)
	for range preparedTasksCount {
		resp := support.MakeCreateTaskRequest(t, app, "Dated task", nil, &date, categoryId, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)

		var task tasksTransport.TaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&task), "failed to decode task response")
		tasks = append(tasks, task)
	}
	return tasks
}

func prepareDoneDatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	tasks := prepareOpenDatedUntrashedTasks(t, app, categoryId, cookies...)
	for index := range tasks {
		resp := support.MakeToggleDoneRequest(t, app, tasks[index].Id, tasks[index].Rank, cookies...)
		require.Equal(t, http.StatusNoContent, resp.Code)
		tasks[index].Done = true
	}
	return tasks
}

func prepareTrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	date, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")
	trashedDate := time.Date(2026, time.March, 26, 10, 30, 0, 0, time.UTC)

	trashTask := func(task tasksTransport.TaskResponse) tasksTransport.TaskResponse {
		resp := support.MakeTrashTaskRequest(t, app, task.Id, trashedDate, cookies...)
		require.Equal(t, http.StatusNoContent, resp.Code)
		task.Rank = nil
		task.TrashedDate = &trashedDate
		return task
	}

	openUndatedResp := support.MakeCreateTaskRequest(t, app, "Open undated task", nil, nil, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, openUndatedResp.Code)
	var openUndatedTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(openUndatedResp.Body).Decode(&openUndatedTask),
		"failed to decode task response",
	)

	openDatedResp := support.MakeCreateTaskRequest(t, app, "Open dated task", nil, &date, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, openDatedResp.Code)
	var openDatedTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(openDatedResp.Body).Decode(&openDatedTask),
		"failed to decode task response",
	)

	doneDatedResp := support.MakeCreateTaskRequest(t, app, "Done dated task", nil, &date, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, doneDatedResp.Code)
	var doneDatedTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(doneDatedResp.Body).Decode(&doneDatedTask),
		"failed to decode task response",
	)
	require.NotNil(t, doneDatedTask.Rank, "expected created task to have a rank")
	toggleDoneResp := support.MakeToggleDoneRequest(t, app, doneDatedTask.Id, doneDatedTask.Rank, cookies...)
	require.Equal(t, http.StatusNoContent, toggleDoneResp.Code)
	doneDatedTask.Done = true

	return []tasksTransport.TaskResponse{
		trashTask(openUndatedTask),
		trashTask(openDatedTask),
		trashTask(doneDatedTask),
	}
}
