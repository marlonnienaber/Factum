package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksGetTrashedTasksExistUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed/exists", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestTasksGetTrashedTasksExistNoTrashedTasks(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	support.PrepareOpenUndatedUntrashedTasks(t, app, category.Id, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed/exists", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksExistResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.False(t, respBody.TrashedTasksExist, "unexpected trashed tasks")
}

func TestTasksGetTrashedTasksExist(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, cookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed/exists", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksExistResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.True(t, respBody.TrashedTasksExist, "expected trashed tasks")
}

func TestTasksGetTrashedTasksExistIgnoresOtherUsersTasks(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email, support.TestUser2Email)
	app := sut.App
	firstSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	firstUserCookies := firstSignUpResp.Result().Cookies()
	secondSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser2Password)
	secondUserCookies := secondSignUpResp.Result().Cookies()

	categoryResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", secondUserCookies...)
	require.Equal(t, http.StatusCreated, categoryResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(categoryResp.Body).Decode(&category), "failed to decode category response")
	support.PrepareTrashedTasks(t, app, category.Id, support.TestTaskDate, secondUserCookies...)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/tasks/trashed/exists", nil, firstUserCookies...)
	require.Equal(t, http.StatusOK, resp.Code)
	var respBody tasksTransport.GetTrashedTasksExistResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.False(t, respBody.TrashedTasksExist, "unexpected trashed tasks owned by another user")
}
