package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/require"
	"roci.dev/fracdex"
)

func TestTasksToggleDoneUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/toggle-done", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksToggleDoneInvalidId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/invalid/toggle-done", nil, signUpResp.Result().Cookies()...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_task_id", "Task id must be a positive integer.")
}

func TestTasksToggleDoneTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email: support.TestUser1Email, Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/toggle-done", nil, signUpResp.Result().Cookies()...)

	support.AssertErrorResponse(t, resp, http.StatusNotFound, "task_not_found", "No task exists for the provided id.")
}

func TestTasksToggleDoneForDatedOpenUntrashedTask(t *testing.T) {
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
	sourceTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := sourceTasks[0]
	lastDestinationTask := destinationTasks[len(destinationTasks)-1]
	require.NotNil(t, lastDestinationTask.Rank, "expected last destination task to have a rank")
	expectedRank, err := fracdex.KeyBetween(*lastDestinationTask.Rank, "")
	require.NoError(t, err, "generating expected rank failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.ToggleTaskDoneResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode toggle task done response")
	require.Equal(t, task.Date, &response.Date)
	require.NotNil(t, response.Rank)
	require.Equal(t, expectedRank, *response.Rank)
}

func TestTasksToggleDoneForDatedDoneUntrashedTask(t *testing.T) {
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
	sourceTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	destinationTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	task := sourceTasks[0]
	lastDestinationTask := destinationTasks[len(destinationTasks)-1]
	require.NotNil(t, lastDestinationTask.Rank, "expected last destination task to have a rank")
	expectedRank, err := fracdex.KeyBetween(*lastDestinationTask.Rank, "")
	require.NoError(t, err, "generating expected rank failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.ToggleTaskDoneResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode toggle task done response")
	require.Equal(t, task.Date, &response.Date)
	require.NotNil(t, response.Rank)
	require.Equal(t, expectedRank, *response.Rank)
}

func TestTasksToggleDoneForUndatedOpenUntrashedTask(t *testing.T) {
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
	today := shared.NewDate(time.Now())
	destinationTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	sourceTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	task := sourceTasks[0]
	lastDestinationTask := destinationTasks[len(destinationTasks)-1]
	require.NotNil(t, lastDestinationTask.Rank, "expected last destination task to have a rank")
	expectedRank, err := fracdex.KeyBetween(*lastDestinationTask.Rank, "")
	require.NoError(t, err, "generating expected rank failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.ToggleTaskDoneResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode toggle task done response")
	require.Equal(t, today, response.Date)
	require.NotNil(t, response.Rank)
	require.Equal(t, expectedRank, *response.Rank)
}

func TestTasksToggleDoneForUndatedOpenTrashedTask(t *testing.T) {
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
	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range tasks {
		if !candidate.Done && candidate.Date == nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open undated trashed task")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.ToggleTaskDoneResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode toggle task done response")
	require.Equal(t, shared.NewDate(time.Now()), response.Date)
	require.Nil(t, response.Rank)
}

func TestTasksToggleDoneForDatedDoneTrashedTask(t *testing.T) {
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
	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range tasks {
		if candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected a done dated trashed task")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.ToggleTaskDoneResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode toggle task done response")
	require.Equal(t, task.Date, &response.Date)
	require.Nil(t, response.Rank)
}

func TestTasksToggleDoneForDatedOpenTrashedTask(t *testing.T) {
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
	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	var task tasksTransport.TaskResponse
	for _, candidate := range tasks {
		if !candidate.Done && candidate.Date != nil {
			task = candidate
			break
		}
	}
	require.NotZero(t, task.Id, "expected an open dated trashed task")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, fmt.Sprintf("/api/tasks/%d/toggle-done", task.Id), nil, cookies...)

	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.ToggleTaskDoneResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode toggle task done response")
	require.Equal(t, task.Date, &response.Date)
	require.Nil(t, response.Rank)
}
