package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesGetCategoriesUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/categories", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestCategoriesGetCategoriesEmpty(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodGet,
		"/api/categories",
		nil,
		signUpResp.Result().Cookies()...,
	)

	require.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, "[]", resp.Body.String())
}

func TestCategoriesGetCategories(t *testing.T) {
	sut := support.PrepareTestEnvironment(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	expectedCategories := make([]categoriesTransport.CategoryResponse, 0, 3)
	for _, category := range []struct {
		title string
		color string
	}{
		{title: "Work", color: "#3b82f6"},
		{title: "Personal", color: "#22c55e"},
		{title: "Shopping", color: "#f97316"},
	} {
		createResp := support.MakeCreateCategoryRequest(t, app, category.title, category.color, cookies...)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var createCategoryResponse categoriesTransport.CreateCategoryResponse
		require.NoError(t, json.NewDecoder(createResp.Body).Decode(&createCategoryResponse), "failed to decode category response")
		expectedCategories = append(expectedCategories, categoriesTransport.CategoryResponse{
			Id:    createCategoryResponse.Id,
			Title: category.title,
			Color: category.color,
		})
	}

	otherSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser2Password)
	otherCreateResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Other user's category",
		"#3b82f6",
		otherSignUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, otherCreateResp.Code)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/categories", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)

	expectedBody, err := json.Marshal(expectedCategories)
	require.NoError(t, err, "failed to encode expected categories")
	assert.JSONEq(t, string(expectedBody), resp.Body.String())
}
