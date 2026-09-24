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

func TestTasksGetDatedUntrashedDoneTasksInvalidCursorDateProvided(t *testing.T) {
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
			"cursorDate": {"invalid"},
			"cursorRank": {"a0"},
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

func TestTasksGetDatedUntrashedDoneTasksInvalidCursorRankProvided(t *testing.T) {
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
			"cursorDate": {"2026-03-25"},
			"cursorRank": {"!"},
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

func TestTasksGetDatedUntrashedDoneTasksInvalidStartDateProvided(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{"startDate": {"invalid"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksInvalidEndDateProvided(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{"endDate": {"invalid"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksStartDateAfterEndDate(t *testing.T) {
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

func TestTasksGetDatedUntrashedDoneTasksEmptyCategoryIdProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done?"+url.Values{"categoryId": {""}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetDatedUntrashedDoneTasksNonIntegerCategoryIdProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done?"+url.Values{"categoryId": {"invalid"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetDatedUntrashedDoneTasksZeroCategoryIdProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done?"+url.Values{"categoryId": {"0"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetDatedUntrashedDoneTasksNegativeCategoryIdProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done?"+url.Values{"categoryId": {"-1"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetDatedUntrashedDoneTasksDuplicateCategoryIdsProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done?"+url.Values{"categoryId": {"1", "1"}}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetDatedUntrashedDoneTasksTooManyCategoryIdsProvided(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{Email: support.TestUser1Email, Password: support.TestUser1Password})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	categoryIds := make([]string, 31)
	for index := range categoryIds {
		categoryIds[index] = strconv.Itoa(index + 1)
	}
	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/dated-untrashed-done?"+url.Values{"categoryId": categoryIds}.Encode(), nil, cookies...)

	support.AssertErrorResponse(t, resp, http.StatusBadRequest, "invalid_query_parameter", "URL contains an invalid query parameter")
}

func TestTasksGetDatedUntrashedDoneTasksCursorDateWithoutCursorRank(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{"cursorDate": {"2026-03-25"}}.Encode(),
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

func TestTasksGetDatedUntrashedDoneTasksCursorRankWithoutCursorDate(t *testing.T) {
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
		"/api/tasks/dated-untrashed-done?"+url.Values{"cursorRank": {"a0"}}.Encode(),
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
			"cursorDate": {cursor.Date.Format(time.DateOnly)},
			"cursorRank": {cursor.Rank},
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

func TestTasksGetDatedUntrashedDoneTasksFiltersByCategoryIds(t *testing.T) {
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

	taskDate := shared.NewDate(time.Now())
	firstCategoryTasks := support.PrepareDoneDatedUntrashedTasks(t, app, firstCategory.Id, taskDate, cookies...)
	secondCategoryTasks := support.PrepareDoneDatedUntrashedTasks(t, app, secondCategory.Id, taskDate, cookies...)
	expectedTasks := append(firstCategoryTasks, secondCategoryTasks...)
	support.PrepareDoneDatedUntrashedTasks(t, app, thirdCategory.Id, taskDate, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done?"+url.Values{
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

func TestTasksGetDatedUntrashedDoneTasksFiltersByStartDate(t *testing.T) {
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
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	futureTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, futureDate, cookies...)
	expectedTasks := append(todayTasks, futureTasks...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done?"+url.Values{"startDate": {today.Format(time.DateOnly)}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedDoneTasksFiltersByEndDate(t *testing.T) {
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
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, futureDate, cookies...)
	expectedTasks := append(pastTasks, todayTasks...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done?"+url.Values{"endDate": {today.Format(time.DateOnly)}}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, expectedTasks, respBody.Tasks, "unexpected tasks")
	assert.False(t, respBody.HasMore, "unexpected additional page")
}

func TestTasksGetDatedUntrashedDoneTasksFiltersByInclusiveDateRange(t *testing.T) {
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
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, pastDate, cookies...)
	todayTasks := support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, today, cookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, category.Id, futureDate, cookies...)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/tasks/dated-untrashed-done?"+url.Values{
			"startDate": {today.Format(time.DateOnly)},
			"endDate":   {today.Format(time.DateOnly)},
		}.Encode(),
		nil,
		cookies...,
	)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetDatedUntrashedTasksResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode task response")

	assert.Equal(t, todayTasks, respBody.Tasks, "unexpected tasks")
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
