package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesDeleteCategoryUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	resp := support.MakeJsonRequest(t, app, http.MethodDelete, "/api/categories/1", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestCategoriesDeleteCategoryInvalidId(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodDelete,
		"/api/categories/invalid",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_category_id",
		"Category id must be a positive integer.",
	)
}

func TestCategoriesDeleteCategoryNotFound(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodDelete,
		"/api/categories/1",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusNotFound,
		"category_not_found",
		"No category exists for the provided id.",
	)
}

func TestCategoriesDeleteCategoryStillReferenced(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
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
		"Referenced task",
		nil,
		nil,
		category.Id,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createTaskResp.Code)

	deleteResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodDelete,
		"/api/categories/1",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		deleteResp,
		http.StatusConflict,
		"category_still_referenced",
		"Category cannot be deleted while it is referenced by tasks.",
	)
}

func TestCategoriesDeleteCategory(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, createCategoryResp.Code)

	deleteResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodDelete,
		"/api/categories/1",
		nil,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusNoContent, deleteResp.Code)

	assert.Empty(t, deleteResp.Body.String(), "expected an empty response body")
}
