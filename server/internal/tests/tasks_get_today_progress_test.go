package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksGetTodayProgressUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/today-progress", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetTodayProgressNoTasks(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/today-progress", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTodayProgressResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Zero(t, respBody.OpenTaskCount)
	assert.Zero(t, respBody.DoneTaskCount)
}

func TestTasksGetTodayProgress(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	firstCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, firstCategoryResp.Code)
	var firstCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(firstCategoryResp.Body).Decode(&firstCategory), "failed to decode category response")
	secondCategoryResp := support.MakeCreateCategoryRequest(t, app, "Personal", "#22c55e", cookies...)
	require.Equal(t, http.StatusCreated, secondCategoryResp.Code)
	var secondCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(secondCategoryResp.Body).Decode(&secondCategory), "failed to decode category response")

	today := shared.NewDate(time.Now())
	openTasks := support.PrepareOpenDatedUntrashedTasks(t, app, firstCategory.Id, today, cookies...)
	doneTasks := support.PrepareDoneDatedUntrashedTasks(t, app, secondCategory.Id, today, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/today-progress", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTodayProgressResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, int64(len(openTasks)), respBody.OpenTaskCount)
	assert.Equal(t, int64(len(doneTasks)), respBody.DoneTaskCount)
}

func TestTasksGetTodayProgressExcludesOtherTasks(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email, support.TestUser2Email)
	app := sut.App
	firstSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	firstUserCookies := firstSignUpResp.Result().Cookies()
	secondSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser2Password)
	secondUserCookies := secondSignUpResp.Result().Cookies()

	firstCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", firstUserCookies...)
	require.Equal(t, http.StatusCreated, firstCategoryResp.Code)
	var firstCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(firstCategoryResp.Body).Decode(&firstCategory), "failed to decode category response")
	secondCategoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", secondUserCookies...)
	require.Equal(t, http.StatusCreated, secondCategoryResp.Code)
	var secondCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(secondCategoryResp.Body).Decode(&secondCategory), "failed to decode category response")

	today := shared.NewDate(time.Now())
	nonToday := shared.NewDate(today.AddDate(0, 0, -1))
	support.PrepareOpenUndatedUntrashedTasks(t, app, firstCategory.Id, firstUserCookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, firstCategory.Id, nonToday, firstUserCookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, firstCategory.Id, nonToday, firstUserCookies...)
	support.PrepareTrashedTasks(t, app, firstCategory.Id, today, firstUserCookies...)
	support.PrepareOpenDatedUntrashedTasks(t, app, secondCategory.Id, today, secondUserCookies...)
	support.PrepareDoneDatedUntrashedTasks(t, app, secondCategory.Id, today, secondUserCookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/today-progress", nil, firstUserCookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTodayProgressResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Zero(t, respBody.OpenTaskCount)
	assert.Zero(t, respBody.DoneTaskCount)
}
