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
		"invalid_drag_and_drop_gesture",
		"The combination of task ID and request body properties does not represent a valid drag and drop gesture.",
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
		"invalid_drag_and_drop_gesture",
		"The combination of task ID and request body properties does not represent a valid drag and drop gesture.",
	)
}

func TestTasksDragAndDropMovingFromDatedOpenUntrashedToDatedDoneUntrashed(t *testing.T) {
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
	openTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(doneTasks[1].Rank, doneTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &openTasks[0].Date,
		Done:           true,
		Trashed:        false,
		PreviousTaskId: &doneTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", openTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromDatedDoneUntrashedToDatedOpenUntrashed(t *testing.T) {
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
	openTasks := support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(openTasks[1].Rank, openTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &doneTasks[0].Date,
		Done:           false,
		Trashed:        false,
		PreviousTaskId: &openTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", doneTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingWithinDatedOpenUntrashed(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &tasks[0].Date,
		Done:           false,
		Trashed:        false,
		PreviousTaskId: &tasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingWithinDatedDoneUntrashed(t *testing.T) {
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
	tasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(tasks[1].Rank, tasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &tasks[0].Date,
		Done:           true,
		Trashed:        false,
		PreviousTaskId: &tasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingWithinUndatedOpenUntrashed(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           nil,
		Done:           false,
		Trashed:        false,
		PreviousTaskId: &tasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromUndatedOpenUntrashedToDatedOpenUntrashed(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &datedTasks[0].Date,
		Done:           false,
		Trashed:        false,
		PreviousTaskId: &datedTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", undatedTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromDatedOpenUntrashedToUndatedOpenUntrashed(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           nil,
		Done:           false,
		Trashed:        false,
		PreviousTaskId: &undatedTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", datedTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromDatedOpenUntrashedToDatedOpenUntrashed(t *testing.T) {
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

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &targetDate,
		Done:           false,
		Trashed:        false,
		PreviousTaskId: nil,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromDatedDoneUntrashedToDatedDoneUntrashed(t *testing.T) {
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
	targetDate, err := shared.ParseDate("2026-03-26")
	require.NoError(t, err, "parsing test date failed")
	targetTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, targetDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(targetTasks[1].Rank, targetTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &targetDate,
		Done:           true,
		Trashed:        false,
		PreviousTaskId: &targetTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", sourceTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromDatedDoneUntrashedToDatedDoneTrashed(t *testing.T) {
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
	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(trashedTasks[1].Rank, trashedTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &sourceTasks[0].Date,
		Done:           true,
		Trashed:        true,
		PreviousTaskId: &trashedTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", sourceTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingFromTrashedToDatedDoneUntrashed(t *testing.T) {
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
	trashedTasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	targetTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(targetTasks[1].Rank, targetTasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           &targetTasks[0].Date,
		Done:           true,
		Trashed:        false,
		PreviousTaskId: &targetTasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", trashedTasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}

func TestTasksDragAndDropMovingWithinTrashed(t *testing.T) {
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
	tasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	expectedRank, err := fracdex.KeyBetween(tasks[1].Rank, tasks[2].Rank)
	require.NoError(t, err, "generating expected rank failed")

	payload, err := json.Marshal(tasksTransport.TaskDragAndDropRequest{
		Date:           nil,
		Done:           false,
		Trashed:        true,
		PreviousTaskId: &tasks[1].Id,
	})
	require.NoError(t, err, "serializing request failed")
	path := fmt.Sprintf("/api/tasks/%d/drag-and-drop", tasks[0].Id)
	resp := support.MakeJsonRequest(t, app, http.MethodPatch, path, payload, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var response tasksTransport.TaskDragAndDropResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode drag and drop response")

	require.Equal(t, expectedRank, response.Rank)
}
