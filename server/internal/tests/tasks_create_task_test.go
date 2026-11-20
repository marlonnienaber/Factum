package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"roci.dev/fracdex"
)

func TestTasksCreateTaskUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:      "Task",
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/tasks", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksCreateTaskNoTitleProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(map[string]any{
		"categoryId": 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"title",
		"title must be at least 1 characters long.",
	)
}

func TestTasksCreateTaskTitleTooShort(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:      "",
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"title",
		"title must be at least 1 characters long.",
	)
}

func TestTasksCreateTaskTitleTooLong(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:      strings.Repeat("a", maximumTaskTitleLength+1),
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"title",
		"title must be at most 100 characters long.",
	)
}

func TestTasksCreateTaskDescriptionTooShort(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	description := ""
	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:       "Task",
		Description: &description,
		CategoryId:  1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"description",
		"description must be at least 1 characters long.",
	)
}

func TestTasksCreateTaskDescriptionTooLong(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	description := strings.Repeat("a", maximumTaskDescriptionLength+1)
	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:       "Task",
		Description: &description,
		CategoryId:  1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"description",
		"description must be at most 500 characters long.",
	)
}

func TestTasksCreateTaskInvalidDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(map[string]any{
		"title":      "Task",
		"date":       "2026-02-30",
		"categoryId": 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"json_parsing_error",
		"Request body does not represent valid JSON.",
	)
}

func TestTasksCreateTaskNoCategoryId(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(map[string]any{
		"title": "Task",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"categoryId",
		"categoryId must be greater than 0.",
	)
}

func TestTasksCreateTaskInvalidCategoryId(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:      "Task",
		CategoryId: -1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"categoryId",
		"categoryId must be greater than 0.",
	)
}

func TestTasksCreateTaskNonExistingCategoryId(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:      "Task",
		CategoryId: 1,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"category_not_found",
		"No category exists for the provided categoryId.",
	)
}

func TestTasksCreateTaskWithCategoryOwnedByDifferentUser(t *testing.T) {
	sut := support.PrepareTestEnvironment(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	app := sut.App
	ownerSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	otherSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser2Password)

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", ownerSignUpResp.Result().Cookies()...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	resp := support.MakeCreateTaskRequest(
		t,
		app,
		"Task",
		nil,
		nil,
		category.Id,
		otherSignUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"category_not_found",
		"No category exists for the provided categoryId.",
	)
}

func TestTasksCreateTaskWithDescriptionWithDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		"#3b82f6",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	description := "Task description"
	date, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")
	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:       "Task",
		Description: &description,
		Date:        &date,
		CategoryId:  category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, resp.Code)
	var respBody tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, int64(1), respBody.Id, "unexpected task id")
	_, err = fracdex.Float64Approx(respBody.Rank)
	assert.NoError(t, err, "expected a valid task rank")
}

func TestTasksCreateTaskWithoutDescriptionWithoutDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		"#3b82f6",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	payload, err := json.Marshal(map[string]any{
		"title":      "Task",
		"categoryId": category.Id,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/tasks",
		payload,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, resp.Code)
	var respBody tasksTransport.CreateTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, int64(1), respBody.Id, "unexpected task id")
	_, err = fracdex.Float64Approx(respBody.Rank)
	assert.NoError(t, err, "expected a valid task rank")
}

func TestTasksCreateTaskMaximumTaskCountReached(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	createCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(createCategoryResp.Body).Decode(&category), "failed to decode category response")

	for range maximumTaskCountPerUser {
		resp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, category.Id, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)
	}

	resp := support.MakeCreateTaskRequest(t, app, "Task", nil, nil, category.Id, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusConflict,
		"maximum_task_count_reached",
		"A maximum of 1000 tasks can be created per user.",
	)
}

const (
	maximumTaskTitleLength       = 100
	maximumTaskDescriptionLength = 500
	maximumTaskCountPerUser      = 1_000
)
