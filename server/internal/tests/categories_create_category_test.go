package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesCreateCategoryUnauthenticated(t *testing.T) {
	app := support.PrepareTestEnvironment(t)

	payload, err := json.Marshal(categoriesTransport.CreateCategoryRequest{
		Title: "Work",
		Color: "#3b82f6",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/categories", payload)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestCategoriesCreateCategoryInvalidColor(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	payload, err := json.Marshal(categoriesTransport.CreateCategoryRequest{
		Title: "Work",
		Color: "#3B82F6",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(t, resp, "color", "color is invalid.")
}

func TestCategoriesCreateCategoryNoTitle(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	payload, err := json.Marshal(map[string]string{"color": "#3b82f6"})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"title",
		"title must be at least 1 characters long.",
	)
}

func TestCategoriesCreateCategoryTitleTooSmall(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	payload, err := json.Marshal(categoriesTransport.CreateCategoryRequest{
		Title: "",
		Color: "#3b82f6",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"title",
		"title must be at least 1 characters long.",
	)
}

func TestCategoriesCreateCategoryTitleTooLong(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	payload, err := json.Marshal(categoriesTransport.CreateCategoryRequest{
		Title: strings.Repeat("a", maximumCategoryTitleLength+1),
		Color: "#3b82f6",
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories",
		payload,
		signUpResp.Result().Cookies()...,
	)

	support.AssertValidationErrorResponse(
		t,
		resp,
		"title",
		"title must be at most 50 characters long.",
	)
}

func TestCategoriesCreateCategoryTitleAlreadyInUse(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	firstResp := support.MakeCreateCategoryRequest(t, app, "Work", "#111111", cookies...)
	require.Equal(t, http.StatusCreated, firstResp.Code)

	secondResp := support.MakeCreateCategoryRequest(t, app, "Work", "#222222", cookies...)

	support.AssertErrorResponse(
		t,
		secondResp,
		http.StatusConflict,
		"category_title_already_in_use",
		"Category title is already in use.",
	)
}

func TestCategoriesCreateCategoryColorAlreadyInUse(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)
	cookies := signUpResp.Result().Cookies()

	firstResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", cookies...)
	require.Equal(t, http.StatusCreated, firstResp.Code)

	secondResp := support.MakeCreateCategoryRequest(t, app, "Personal", "#3b82f6", cookies...)

	support.AssertErrorResponse(
		t,
		secondResp,
		http.StatusConflict,
		"category_color_already_in_use",
		"Category color is already in use.",
	)
}

func TestCategoriesCreateCategoryTitleCanBeReusedByDifferentUser(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	firstSignUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, firstSignUpResp.Code)
	secondSignUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser2Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, secondSignUpResp.Code)

	firstResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", firstSignUpResp.Result().Cookies()...)
	require.Equal(t, http.StatusCreated, firstResp.Code)
	secondResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", secondSignUpResp.Result().Cookies()...)
	require.Equal(t, http.StatusCreated, secondResp.Code)

	var firstCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(firstResp.Body).Decode(&firstCategory), "failed to decode first category response")
	var secondCategory categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(secondResp.Body).Decode(&secondCategory), "failed to decode second category response")

	assert.Equal(t, firstCategory.Title, secondCategory.Title, "expected users to share a category title")
	assert.NotEqual(t, firstCategory.Id, secondCategory.Id, "expected separate categories")
}

func TestCategoriesCreateCategory(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUser1Email)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUser1Email,
		Password: support.TestUser1Password,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	title := "Work & personal"
	color := "#12abef"
	payload, err := json.Marshal(categoriesTransport.CreateCategoryRequest{
		Title: title,
		Color: color,
	})
	require.NoError(t, err, "serializing request failed")
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories",
		payload,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, resp.Code)

	var respBody categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, int64(1), respBody.Id, "unexpected category id")
	assert.Equal(t, title, respBody.Title, "unexpected category title")
	assert.Equal(t, color, respBody.Color, "unexpected category color")
}

const maximumCategoryTitleLength = 50
