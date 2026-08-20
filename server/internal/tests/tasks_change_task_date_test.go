package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"roci.dev/fracdex"
)

func TestTasksChangeDateUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   1,
		Rank: "a0",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksChangeDateInvalidId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   -1,
		Rank: "a0",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "id", "id must be greater than 0.")
}

func TestTasksChangeDateInvalidDate(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(map[string]any{
		"id":   1,
		"date": "2026-02-30",
		"rank": "a0",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"json_parsing_error",
		"Request body does not represent valid JSON.",
	)
}

func TestTasksChangeDateInvalidRank(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   1,
		Rank: "!",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "rank", "rank must be a valid fractional rank.")
}

func TestTasksChangeDateTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	targetDate, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")
	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   1,
		Date: &targetDate,
		Rank: "a0",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"task_not_found",
		"No task exists for the provided id.",
	)
}

func TestTasksChangeDateRankAlreadyExists(t *testing.T) {
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
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := undatedTasks[0]
	require.NotNil(t, datedTasks[0].Date, "expected dated task to have a date")
	require.NotNil(t, datedTasks[0].Rank, "expected dated task to have a rank")

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   task.Id,
		Date: datedTasks[0].Date,
		Rank: *datedTasks[0].Rank,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusConflict,
		"task_rank_already_exists",
		"The provided rank is already in use.",
	)
}

func TestTasksChangeDateTrashedTaskDateToDate(t *testing.T) {
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
	var task tasksTransport.TaskResponse
	for _, candidate := range trashedTasks {
		if !candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open dated trashed task")

	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")
	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   task.Id,
		Date: &targetDate,
		Rank: "a0",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"task_date_change_not_allowed",
		"This date transition is not allowed.",
	)
}

func TestTasksChangeDateUndatedOpenUntrashedToDatedOpenUntrashedTask(t *testing.T) {
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
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := undatedTasks[0]
	lastDatedTask := datedTasks[len(datedTasks)-1]
	require.NotNil(t, lastDatedTask.Date, "expected dated task to have a date")
	require.NotNil(t, lastDatedTask.Rank, "expected dated task to have a rank")
	rank, err := fracdex.KeyBetween(*lastDatedTask.Rank, "")
	require.NoError(t, err, "generating task rank failed")

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   task.Id,
		Date: lastDatedTask.Date,
		Rank: rank,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, lastDatedTask.Date, respBody.Date, "unexpected task date")
	assert.Equal(t, &rank, respBody.Rank, "unexpected task rank")
	assert.False(t, respBody.Done, "expected task to remain open")
}

func TestTasksChangeDateDatedOpenUntrashedToUndatedOpenUntrashedTask(t *testing.T) {
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

	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	undatedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	task := datedTasks[0]
	lastUndatedTask := undatedTasks[len(undatedTasks)-1]
	require.NotNil(t, lastUndatedTask.Rank, "expected undated task to have a rank")
	rank, err := fracdex.KeyBetween(*lastUndatedTask.Rank, "")
	require.NoError(t, err, "generating task rank failed")

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   task.Id,
		Rank: rank,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Nil(t, respBody.Date, "expected task date to be removed")
	assert.Equal(t, &rank, respBody.Rank, "unexpected task rank")
	assert.False(t, respBody.Done, "expected task to remain open")
}

func TestTasksChangeDateDatedOpenUntrashedToDatedOpenUntrashedTask(t *testing.T) {
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

	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := datedTasks[0]
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")
	rank, err := fracdex.KeyBetween("", "")
	require.NoError(t, err, "generating task rank failed")

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   task.Id,
		Date: &targetDate,
		Rank: rank,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.Equal(t, &rank, respBody.Rank, "unexpected task rank")
	assert.False(t, respBody.Done, "expected task to remain open")
}

func TestTasksChangeDateDatedDoneUntrashedToDatedDoneUntrashedTask(t *testing.T) {
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

	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, cookies...)
	task := doneTasks[0]
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")
	rank, err := fracdex.KeyBetween("", "")
	require.NoError(t, err, "generating task rank failed")

	payload, err := json.Marshal(tasksTransport.ChangeTaskDateRequest{
		Id:   task.Id,
		Date: &targetDate,
		Rank: rank,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/change-date", payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, &targetDate, respBody.Date, "unexpected task date")
	assert.Equal(t, &rank, respBody.Rank, "unexpected task rank")
	assert.True(t, respBody.Done, "expected task to remain done")
}
