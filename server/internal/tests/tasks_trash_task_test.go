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

func TestTasksTrashTaskUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/trash?"+url.Values{"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)}}.Encode(),
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

func TestTasksTrashTaskInvalidId(t *testing.T) {
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
		http.MethodPatch,
		"/api/tasks/invalid/trash?"+url.Values{"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)}}.Encode(),
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

func TestTasksTrashTaskNotFound(t *testing.T) {
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
		http.MethodPatch,
		"/api/tasks/1/trash?"+url.Values{"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)}}.Encode(),
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

func TestTasksTrashTaskAlreadyTrashed(t *testing.T) {
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

	tasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	secondTrashResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf("/api/tasks/%d/trash", tasks[0].Id),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		secondTrashResp,
		http.StatusBadRequest,
		"task_already_trashed",
		"Task is already trashed.",
	)
}

func TestTasksTrashTask(t *testing.T) {
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

	existingTrashedTasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	tasks := support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf("/api/tasks/%d/trash", tasks[0].Id),
		nil,
		cookies...,
	)

	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.TrashTaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode trash task response")
	assert.NotEmpty(t, respBody.Rank, "expected trashed task to have a rank")
	assert.Less(t, respBody.Rank, existingTrashedTasks[0].Rank, "expected task to be inserted at the beginning of trash")
}
