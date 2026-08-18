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

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/invalid/trash?"+url.Values{"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)}}.Encode(),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_task_id",
		"Task id must be a positive integer.",
	)
}

func TestTasksTrashTaskNoTrashedDate(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/trash",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"missing_query_parameter",
		"URL is missing a required query parameter.",
	)
}

func TestTasksTrashTaskInvalidTrashedDate(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/trash?"+url.Values{"trashedDate": {"invalid"}}.Encode(),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter",
		"URL contains an invalid query parameter",
	)
}

func TestTasksTrashTaskNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		"/api/tasks/1/trash?"+url.Values{"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)}}.Encode(),
		nil,
		signUpResp.Result().Cookies()...,
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

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	createTaskResp := support.MakeCreateTaskRequest(
		t,
		app,
		"Task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createTaskResp.Code)

	var task tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(createTaskResp.Body).Decode(&task),
		"failed to decode task response",
	)

	trashedDate := time.Now().UTC()
	firstTrashResp := support.MakeTrashTaskRequest(
		t,
		app,
		task.Id,
		trashedDate,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusNoContent, firstTrashResp.Code)

	secondTrashResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/trash?%s",
			task.Id,
			url.Values{"trashedDate": {trashedDate.Add(time.Second).Format(time.RFC3339Nano)}}.Encode(),
		),
		nil,
		signUpResp.Result().Cookies()...,
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

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	var category categoriesTransport.CategoryResponse
	require.NoError(
		t,
		json.NewDecoder(createCategoryResp.Body).Decode(&category),
		"failed to decode category response",
	)

	createTaskResp := support.MakeCreateTaskRequest(
		t,
		app,
		"Task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createTaskResp.Code)

	var task tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(createTaskResp.Body).Decode(&task),
		"failed to decode task response",
	)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPatch,
		fmt.Sprintf(
			"/api/tasks/%d/trash?%s",
			task.Id,
			url.Values{"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)}}.Encode(),
		),
		nil,
		signUpResp.Result().Cookies()...,
	)

	require.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Body.String(), "expected an empty response body")
}
