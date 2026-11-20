package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/tests/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesCreateCategoryUnauthenticated(t *testing.T) {
	sut := support.PrepareTestEnvironment(t)
	app := sut.App

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
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
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
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

func TestCategoriesCreateCategoryMaximumCategoryCountReached(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	cookies := signUpResp.Result().Cookies()

	for index := range maximumCategoryCountPerUser {
		resp := support.MakeCreateCategoryRequest(
			t,
			app,
			fmt.Sprintf("Category %d", index+1),
			fmt.Sprintf("#%06x", index+1),
			cookies...,
		)
		require.Equal(t, http.StatusCreated, resp.Code)
	}

	resp := support.MakeCreateCategoryRequest(t, app, "One too many", "#ffffff", cookies...)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusConflict,
		"maximum_category_count_reached",
		"A maximum of 30 categories can be created per user.",
	)
}

func TestCategoriesCreateCategoryTitleCanBeReusedByDifferentUser(t *testing.T) {
	sut := support.PrepareTestEnvironment(
		t,
		support.TestUser1Email,
		support.TestUser2Email,
	)
	app := sut.App
	firstSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)
	secondSignUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser2Email, support.TestUser1Password)

	firstResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", firstSignUpResp.Result().Cookies()...)
	require.Equal(t, http.StatusCreated, firstResp.Code)
	secondResp := support.MakeCreateCategoryRequest(t, app, "Work", "#3b82f6", secondSignUpResp.Result().Cookies()...)
	require.Equal(t, http.StatusCreated, secondResp.Code)

	var firstCategory categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(firstResp.Body).Decode(&firstCategory), "failed to decode first category response")
	var secondCategory categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(secondResp.Body).Decode(&secondCategory), "failed to decode second category response")

	assert.NotEqual(t, firstCategory.Id, secondCategory.Id, "expected separate categories")
}

func TestCategoriesCreateCategory(t *testing.T) {
	sut := support.PrepareTestEnvironment(t, support.TestUser1Email)
	app := sut.App
	signUpResp := support.MakeVerifyEmailSignUpAndSignInRequests(t, sut, support.TestUser1Email, support.TestUser1Password)

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

	var respBody categoriesTransport.CreateCategoryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, int64(1), respBody.Id, "unexpected category id")
}

const (
	maximumCategoryTitleLength  = 50
	maximumCategoryCountPerUser = 30
)
