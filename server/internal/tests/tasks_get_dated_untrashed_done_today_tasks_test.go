package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

func TestTasksGetDatedUntrashedDoneTodayTasksUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done-today", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetDatedUntrashedDoneTodayTasksInvalidRankProvided(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done-today?"+url.Values{"rank": {"!"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTodayTasksInvalidCategoryIdProvided(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done-today?"+url.Values{"categoryId": {"invalid"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTodayTasksTooManyCategoryIdsProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryIds := make([]string, 31)
	for index := range categoryIds {
		categoryIds[index] = strconv.Itoa(index + 1)
	}
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done-today?"+url.Values{"categoryId": categoryIds}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTodayTasksWithoutCursor(t *testing.T) {
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
	nonToday := shared.NewDate(today.AddDate(0, 0, -1))
	expectedTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, nonToday, cookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done-today", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedDoneTodayTasksWithCursor(t *testing.T) {
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
	tasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	cursor := tasks[0]
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done-today?"+url.Values{"rank": {cursor.Rank}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, tasks[1:], respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedDoneTodayTasksFiltersByCategoryIds(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	firstCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, firstCategoryResp.Code)
	var firstCategory categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(firstCategoryResp.Body).Decode(&firstCategory), "failed to decode first category response")

	secondCategoryResp := support.MakeCreateCategoryRequest(t, app, "Personal", "#22c55e", cookies...)
	require.Equal(t, http.StatusCreated, secondCategoryResp.Code)
	var secondCategory categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(secondCategoryResp.Body).Decode(&secondCategory), "failed to decode second category response")

	thirdCategoryResp := support.MakeCreateCategoryRequest(t, app, "Errands", "#ef4444", cookies...)
	require.Equal(t, http.StatusCreated, thirdCategoryResp.Code)
	var thirdCategory categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(thirdCategoryResp.Body).Decode(&thirdCategory), "failed to decode third category response")

	today := shared.NewDate(time.Now())
	firstCategoryTasks := support.PrepareDoneDatedUntrashedTasks(t, app, firstCategory.Id, today, cookies...)
	secondCategoryTasks := support.PrepareDoneDatedUntrashedTasks(t, app, secondCategory.Id, today, cookies...)
	expectedTasks := append(firstCategoryTasks, secondCategoryTasks...)
	support.PrepareDoneDatedUntrashedTasks(t, app, thirdCategory.Id, today, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done-today?"+url.Values{
			"categoryId": {
				strconv.FormatInt(firstCategory.Id, 10),
				strconv.FormatInt(secondCategory.Id, 10),
			},
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

func TestTasksGetDatedUntrashedDoneTodayTasksHasMore(t *testing.T) {
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
	today := shared.NewDate(time.Now())
	expectedTasks := make([]tasksTransport.DatedUntrashedTaskResponse, 0, taskBatchSize)
	for index := range taskBatchSize + 1 {
		title := fmt.Sprintf("Task %d", index+1)
		createTaskResp := support.MakeCreateTaskRequest(t, app, title, nil, &today, category.Id, cookies...)
		require.Equal(t, http.StatusCreated, createTaskResp.Code)
		var task tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(createTaskResp.Body).Decode(&task), "failed to decode task response")
		toggleDoneResp := support.MakeToggleDoneRequest(t, app, task.Id, cookies...)
		require.Equal(t, http.StatusOK, toggleDoneResp.Code)
		var toggledTask tasksTransport.ToggleTaskDoneResponse
		require.NoError(t, json.NewDecoder(toggleDoneResp.Body).Decode(&toggledTask), "failed to decode toggle task response")
		if index < taskBatchSize {
			expectedTasks = append(expectedTasks, tasksTransport.DatedUntrashedTaskResponse{
				Id:         task.Id,
				CategoryId: category.Id,
				Title:      title,
				Date:       today,
				Rank:       *toggledTask.Rank,
			})
		}
	}

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done-today", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.True(t, respBody.HasMore, "expected additional tasks")
}
