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

func TestCategoriesGetCategoriesUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

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
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	expectedCategories := make([]categoriesTransport.CategoryResponse, 0, 3)
	for _, title := range []string{"Work", "Personal", "Shopping"} {
		createResp := support.MakeCreateCategoryRequest(t, app, title, cookies...)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var category categoriesTransport.CategoryResponse
		require.NoError(t, json.NewDecoder(createResp.Body).Decode(&category), "failed to decode category response")
		expectedCategories = append(expectedCategories, category)
	}

	otherSignUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser2Email,
		Password: support.TestUser2Password,
	})
	require.Equal(t, http.StatusCreated, otherSignUpResp.Code)
	otherCreateResp := support.MakeCreateCategoryRequest(
		t,
		app,
		"Other user's category",
		otherSignUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, otherCreateResp.Code)

	resp := support.MakeJsonRequest(t, app, http.MethodGet, "/api/categories", nil, cookies...)
	require.Equal(t, http.StatusOK, resp.Code)

	expectedBody, err := json.Marshal(expectedCategories)
	require.NoError(t, err, "failed to encode expected categories")
	assert.JSONEq(t, string(expectedBody), resp.Body.String())
}
