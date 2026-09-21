package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
		"description must be at most 500 characters long.",
	)
}

func TestTasksUpdateTaskNoCategoryId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
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

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
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

func TestTasksUpdateTaskWithCategoryOwnedByDifferentUser(t *testing.T) {
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

	firstCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", firstUserCookies...)
	require.Equal(t, http.StatusCreated, firstCategoryResp.Code)
	var firstCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(firstCategoryResp.Body).Decode(&firstCategory), "failed to decode first category response")

	secondCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Personal",
		"#3b82f6",
		secondSignUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, secondCategoryResp.Code)
	var secondCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(secondCategoryResp.Body).Decode(&secondCategory), "failed to decode second category response")

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, firstCategory.Id, firstUserCookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: secondCategory.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, firstUserCookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"category_not_found",
		"No category exists for the provided categoryId.",
	)
}

func TestTasksUpdateTaskChangeTitle(t *testing.T) {
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

	taskResp := support.MakeCreateTaskRequest(t, app, "Old title", nil, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "New title",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, "New title", respBody.Title, "unexpected task title")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskAddDescription(t *testing.T) {
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

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	description := "Description"
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:          task.Id,
		Title:       "Task",
		Description: &description,
		CategoryId:  category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &description, respBody.Description, "unexpected task description")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskRemoveDescription(t *testing.T) {
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

	description := "Description"
	taskResp := support.MakeCreateTaskRequest(t, app, "Task", &description, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Description, "expected task description to be removed")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskChangeDescription(t *testing.T) {
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

	oldDescription := "Old description"
	taskResp := support.MakeCreateTaskRequest(t, app, "Task", &oldDescription, nil, category.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	newDescription := "New description"
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:          task.Id,
		Title:       "Task",
		Description: &newDescription,
		CategoryId:  category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &newDescription, respBody.Description, "unexpected task description")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}

func TestTasksUpdateTaskFromOpenWithoutDateUntrashedToDate(t *testing.T) {
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
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := undatedTasks[0]
	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		Date:       &datedTasks[0].Date,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &datedTasks[0].Date, respBody.Date, "unexpected task date")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromOpenWithDateUntrashedToNoDate(t *testing.T) {
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

	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := datedTasks[0]

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromDoneWithDateUntrashedToNoDate(t *testing.T) {
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

	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := doneTasks[0]

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to become open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromDoneWithDateTrashedToNoDateTrashed(t *testing.T) {
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

	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TrashedTaskResponse
	for _, candidate := range trashedTasks {
		if candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected a done dated trashed task")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to become open")
	assert.Nil(t, respBody.Rank, "expected trashed task not to have a rank")
}

func TestTasksUpdateTaskFromOpenWithDateTrashedToNoDateTrashed(t *testing.T) {
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

	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TrashedTaskResponse
	for _, candidate := range trashedTasks {
		if !candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open dated trashed task")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.Nil(t, respBody.Rank, "expected trashed task not to have a rank")
}

func TestTasksUpdateTaskFromOpenWithoutDateTrashedToDateTrashed(t *testing.T) {
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

	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TrashedTaskResponse
	for _, candidate := range trashedTasks {
		if !candidate.Done && candidate.Date == nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open undated trashed task")
	targetDate, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		Date:       &targetDate,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.Nil(t, respBody.Rank, "expected trashed task not to have a rank")
}

func TestTasksUpdateTaskFromOpenWithDateToOtherDate(t *testing.T) {
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

	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := datedTasks[0]
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		Date:       &targetDate,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.False(t, respBody.Done, "expected task to remain open")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskFromDoneWithDateToOtherDate(t *testing.T) {
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

	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := doneTasks[0]
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		Date:       &targetDate,
		CategoryId: category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.True(t, respBody.Done, "expected task to remain done")
	assert.NotNil(t, respBody.Rank, "expected task to have a rank")
}

func TestTasksUpdateTaskChangeCategory(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	oldCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, oldCategoryResp.Code)
	var oldCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(oldCategoryResp.Body).Decode(&oldCategory), "failed to decode category response")

	newCategoryResp := support.MakeCreateCategoryRequest(t, app, "Personal", "#22c55e", cookies...)
	require.Equal(t, http.StatusCreated, newCategoryResp.Code)
	var newCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(newCategoryResp.Body).Decode(&newCategory), "failed to decode category response")

	taskResp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, oldCategory.Id, cookies...)
	require.Equal(t, http.StatusCreated, taskResp.Code)
	var task tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(taskResp.Body).Decode(&task), "failed to decode task response")

	payload, err := json.Marshal(tasksTransport.UpdateTaskRequest{
		Id:         task.Id,
		Title:      "Task",
		CategoryId: newCategory.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPut, "/api/tasks", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.UpdateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, newCategory.Id, respBody.CategoryId, "unexpected category id")
	assert.Equal(t, task.Rank, respBody.Rank, "expected task rank to remain unchanged")
}
