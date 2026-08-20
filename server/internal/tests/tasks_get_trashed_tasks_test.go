package tests

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksGetTrashedTasksUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetTrashedTasksInvalidTrashedDate(t *testing.T) {
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
		"/api/tasks/trashed?"+url.Values{"trashedDate": {"invalid"}}.Encode(),
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

func TestTasksGetTrashedTasksInvalidId(t *testing.T) {
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
		"/api/tasks/trashed?"+url.Values{"id": {"invalid"}}.Encode(),
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

func TestTasksGetTrashedTasksIncompleteCursor(t *testing.T) {
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
		"/api/tasks/trashed?"+url.Values{
			"trashedDate": {time.Now().UTC().Format(time.RFC3339Nano)},
		}.Encode(),
		nil,
		cookies...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter_combination",
		"URL contains an invalid query parameter combination.",
	)
}

func TestTasksGetTrashedTasksNoCursor(t *testing.T) {
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

	expectedTasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody []tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody, "unexpected tasks")
}

func TestTasksGetTrashedTasksCursor(t *testing.T) {
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

	tasks := support.PrepareTrashedTasks(t, app, category.Id, cookies...)
	cursor := tasks[0]
	require.NotNil(t, cursor.TrashedDate, "expected cursor task to have a trashed date")

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{
			"trashedDate": {cursor.TrashedDate.Format(time.RFC3339Nano)},
			"id":          {strconv.FormatInt(cursor.Id, 10)},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody []tasksTransport.TaskResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, tasks[1:], respBody, "unexpected tasks")
}
