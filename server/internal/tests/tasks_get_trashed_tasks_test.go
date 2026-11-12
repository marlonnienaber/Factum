package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksGetTrashedTasksUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetTrashedTasksInvalidRank(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{"rank": {"!"}}.Encode(),
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

func TestTasksGetTrashedTasksInvalidStartDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{"startDate": {"invalid"}}.Encode(),
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

func TestTasksGetTrashedTasksInvalidEndDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{"endDate": {"invalid"}}.Encode(),
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

func TestTasksGetTrashedTasksStartDateAfterEndDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{
			"startDate": {"2026-03-26"},
			"endDate":   {"2026-03-25"},
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

func TestTasksGetTrashedTasksEmptyCategoryIdProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed?"+url.Values{"categoryId": {""}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetTrashedTasksNonIntegerCategoryIdProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed?"+url.Values{"categoryId": {"invalid"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetTrashedTasksZeroCategoryIdProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed?"+url.Values{"categoryId": {"0"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetTrashedTasksNegativeCategoryIdProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed?"+url.Values{"categoryId": {"-1"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetTrashedTasksDuplicateCategoryIdsProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed?"+url.Values{"categoryId": {"1", "1"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetTrashedTasksTooManyCategoryIdsProvided(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryIds := make([]string, 31)
	for index := range categoryIds {
		categoryIds[index] = strconv.Itoa(index + 1)
	}
	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed?"+url.Values{"categoryId": categoryIds}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetTrashedTasksFiltersByCategoryIds(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
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

	firstCategoryTasks := support.PrepareTrashedTasks(t, app, firstCategory.Id, support.TestTaskDate, cookies...)
	secondCategoryTasks := support.PrepareTrashedTasks(t, app, secondCategory.Id, support.TestTaskDate, cookies...)
	expectedTasks := append(secondCategoryTasks, firstCategoryTasks...)
	support.PrepareTrashedTasks(t, app, thirdCategory.Id, support.TestTaskDate, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{
			"categoryId": {
				strconv.FormatInt(firstCategory.Id, 10),
				strconv.FormatInt(secondCategory.Id, 10),
			},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetTrashedTasksFiltersByStartDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	today := shared.NewDate(time.Now())
	pastDate := shared.NewDate(today.AddDate(0, 0, -1))
	futureDate := shared.NewDate(today.AddDate(0, 0, 1))
	support.PrepareTrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareTrashedTasks(t, app, category.Id, today, cookies...)
	futureTasks := support.PrepareTrashedTasks(t, app, category.Id, futureDate, cookies...)
	expectedTasks := append(futureTasks[1:], todayTasks[1:]...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{"startDate": {today.Format(time.DateOnly)}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetTrashedTasksFiltersByEndDate(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	today := shared.NewDate(time.Now())
	pastDate := shared.NewDate(today.AddDate(0, 0, -1))
	futureDate := shared.NewDate(today.AddDate(0, 0, 1))
	pastTasks := support.PrepareTrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareTrashedTasks(t, app, category.Id, today, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, futureDate, cookies...)
	expectedTasks := append(todayTasks[1:], pastTasks[1:]...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{"endDate": {today.Format(time.DateOnly)}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetTrashedTasksFiltersByInclusiveDateRange(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	today := shared.NewDate(time.Now())
	pastDate := shared.NewDate(today.AddDate(0, 0, -1))
	futureDate := shared.NewDate(today.AddDate(0, 0, 1))
	support.PrepareTrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareTrashedTasks(t, app, category.Id, today, cookies...)
	support.PrepareTrashedTasks(t, app, category.Id, futureDate, cookies...)
	expectedTasks := todayTasks[1:]

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{
			"startDate": {today.Format(time.DateOnly)},
			"endDate":   {today.Format(time.DateOnly)},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetTrashedTasksNoCursor(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	expectedTasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "expected no additional tasks")
}

func TestTasksGetTrashedTasksCursor(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	tasks := support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/trashed?"+url.Values{
			"rank": {tasks[0].Rank},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, tasks[1:], respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "expected no additional tasks")
}

func TestTasksGetTrashedTasksHasMore(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")

	const taskBatchSize = 50
	trashedDate := time.Date(2026, time.March, 26, 10, 30, 0, 0, time.UTC)
	createdTasks := make([]tasksTransport.TrashedTaskResponse, 0, taskBatchSize+1)
	for index := range taskBatchSize + 1 {
		title := fmt.Sprintf("Task %d", index+1)
		createTaskResp := support.MakeCreateTaskRequest(t, app, title, nil, nil, category.Id, cookies...)
		require.Equal(t, http.StatusCreated, createTaskResp.Code)
		var task tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(createTaskResp.Body).Decode(&task), "failed to decode task response")

		trashTaskResp := support.MakeTrashTaskRequest(t, app, task.Id, trashedDate, cookies...)
		require.Equal(t, http.StatusOK, trashTaskResp.Code)
		var trashTask tasksTransport.TrashTaskResponse
		require.NoError(t, json.NewDecoder(trashTaskResp.Body).Decode(&trashTask), "failed to decode trash task response")
		createdTasks = append(createdTasks, tasksTransport.TrashedTaskResponse{
			Id:         task.Id,
			CategoryId: category.Id,
			Title:      title,
			Rank:       trashTask.Rank,
		})
	}

	expectedTasks := make([]tasksTransport.TrashedTaskResponse, 0, taskBatchSize)
	for index := taskBatchSize; index > 0; index-- {
		expectedTasks = append(expectedTasks, createdTasks[index])
	}

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.True(t, respBody.HasMore, "expected additional tasks")
}
