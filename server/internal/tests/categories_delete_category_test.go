package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesDeleteCategoryUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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
		"No category of the logged in user exists with the provided id.",
	)
}

func TestCategoriesDeleteCategoryStillReferenced(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		"#3b82f6",
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

func TestCategoriesDeleteCategoryOwnedByDifferentUser(t *testing.T) {
	sut := support.PrepareTestEnvironment(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	app := sut.App
	ownerSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	otherSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser2Password)

	createResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", ownerSignUpResp.Result().Cookies()...)
	require.Equal(t, http.StatusCreated, createResp.Code)
	var category categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&category), "failed to decode category response")

	deleteResp := support.MakeJsonRequest(
		t,
		app,
		http.MethodDelete,
		fmt.Sprintf("/api/categories/%d", category.Id),
		nil,
		otherSignUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		deleteResp,
		http.StatusNotFound,
		"category_not_found",
		"No category of the logged in user exists with the provided id.",
	)
}

func TestCategoriesDeleteCategory(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	createCategoryResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Work",
		"#3b82f6",
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
