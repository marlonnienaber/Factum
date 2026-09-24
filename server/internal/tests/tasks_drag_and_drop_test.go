package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/require"
	"roci.dev/fracdex"
)

func TestTasksDragAndDropUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)
	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{})
	require.NoError(t, err, "serializing request failed")

	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/drag-and-drop", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksDragAndDropInvalidTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/-1/drag-and-drop", payload, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_task_id", "Task id must be a positive integer.")
}

func TestTasksDragAndDropInvalidPreviousTaskId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	previousTaskId := int64(-1)
	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{PreviousTaskId: &previousTaskId})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/drag-and-drop", payload, cookies...)

	support.AssertValidationErrorResponse(t, resp, "previousTaskId", "previousTaskId must be greater than 0.")
}

func TestTasksDragAndDropInvalidDate(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(map[string]any{"date": "2026-02-30"})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/1/drag-and-drop", payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"json_parsing_error",
		"Request body does not represent valid JSON.",
	)
}

func TestTasksDragAndDropTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, "/api/tasks/999/drag-and-drop", payload, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusNotFound, "task_not_found", "No task exists for the provided id.")
}

func TestTasksDragAndDropPreviousTaskNotFound(t *testing.T) {
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

	missingPreviousTaskId := int64(999)
	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{PreviousTaskId: &missingPreviousTaskId})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"previous_task_not_found",
		"No previous task exists for the provided previousTaskId.",
	)
}

func TestTasksDragAndDropTaskCannotBeItsOwnPreviousTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	tasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{PreviousTaskId: &tasks[0].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_date_and_previous_task_id_combination",
		"The provided date and previousTaskId do not represent an allowed drag and drop gesture.",
	)
}

func TestTasksDragAndDropPreviousTaskBelongsToAnotherPartition(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	undatedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{PreviousTaskId: &datedTasks[0].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", undatedTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_date_and_previous_task_id_combination",
		"The provided date and previousTaskId do not represent an allowed drag and drop gesture.",
	)
}

func TestTasksDragAndDropCompletedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	tasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{Date: &tasks[0].Date, PreviousTaskId: &tasks[1].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_date_and_previous_task_id_combination",
		"The provided date and previousTaskId do not represent an allowed drag and drop gesture.",
	)
}

func TestTasksDragAndDropTrashedTask(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	tasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{Date: tasks[1].Date})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[1].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_date_and_previous_task_id_combination",
		"The provided date and previousTaskId do not represent an allowed drag and drop gesture.",
	)
}

func TestTasksDragAndDropWithinUndatedPartition(t *testing.T) {
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
	expectedRank, err := fracdex.KeyBetween(tasks[1].Rank, tasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{PreviousTaskId: &tasks[1].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropWithinDatedPartition(t *testing.T) {
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
	tasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(tasks[1].Rank, tasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{Date: &tasks[0].Date, PreviousTaskId: &tasks[1].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropFromUndatedToDatedPartition(t *testing.T) {
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
	undatedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(datedTasks[1].Rank, datedTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{Date: &datedTasks[0].Date, PreviousTaskId: &datedTasks[1].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", undatedTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropFromDatedToUndatedPartition(t *testing.T) {
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
	datedTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	undatedTasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	expectedRank, err := fracdex.KeyBetween(undatedTasks[1].Rank, undatedTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{PreviousTaskId: &undatedTasks[1].Id})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", datedTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropFromDatedToDifferentDatedPartition(t *testing.T) {
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
	tasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")
	expectedRank, err := fracdex.KeyBetween("", "")
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{Date: &targetDate})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}
