package tests

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksGetDatedUntrashedDoneTasksUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetDatedUntrashedDoneTasksInvalidDateProvided(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{
			"date": {"invalid"},
			"rank": {"a0"},
		}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksInvalidRankProvided(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{
			"date": {"2026-03-25"},
			"rank": {"!"},
		}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksDateWithoutRank(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{"date": {"2026-03-25"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksRankWithoutDate(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{"rank": {"a0"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksWithoutCursor(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	today := shared.NewDate(time.Now())
	pastDate := shared.NewDate(today.AddDate(0, 0, -1))
	futureDate := shared.NewDate(today.AddDate(0, 0, 1))
	pastTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	futureTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, futureDate, cookies...)
	expectedTasks := append(append(pastTasks, todayTasks...), futureTasks...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, pastDate, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedDoneTasksWithCursor(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	today := shared.NewDate(time.Now())
	pastDate := shared.NewDate(today.AddDate(0, 0, -1))
	futureDate := shared.NewDate(today.AddDate(0, 0, 1))
	pastTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	futureTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, futureDate, cookies...)
	cursor := pastTasks[0]
	expectedTasks := append(append(pastTasks[1:], todayTasks...), futureTasks...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, pastDate, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done?"+url.Values{
			"date": {cursor.Date.Format(time.DateOnly)},
			"rank": {cursor.Rank},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedDoneTasksHasMore(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	const taskBatchSize = 50
	taskDate := shared.NewDate(time.Now())
	allTasks := make([]tasksTransport.DatedUntrashedTaskResponse, 0, taskBatchSize+1)
	for range 17 {
		allTasks = append(allTasks, support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, taskDate, cookies...)...)
	}

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, allTasks[:taskBatchSize], respBody.Tasks, "unexpected tasks")
	assert.True(t, respBody.HasMore, "expected additional tasks")
}
