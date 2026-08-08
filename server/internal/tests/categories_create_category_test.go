package tests

import (
	"encoding/json"
	"net/http"
	"net/url"
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

	resp := support.MakeJsonRequest(t, app, http.MethodPost, "/api/categories?title=Work", nil)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusUnauthorized,
		"authentication_required",
		"This endpoint requires authentication. Please sign-in.",
	)
}

func TestCategoriesCreateCategoryNoTitle(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"missing_query_parameter",
		"URL is missing a required query parameter.",
	)
}

func TestCategoriesCreateCategoryTitleTooSmall(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories?title=",
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter",
		"URL contains an invalid query parameter",
	)
}

func TestCategoriesCreateCategoryTitleTooLong(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories?title="+strings.Repeat("a", maximumCategoryTitleLength+1),
		nil,
		signUpResp.Result().Cookies()...,
	)

	support.AssertErrorResponse(
		t,
		resp,
		http.StatusBadRequest,
		"invalid_query_parameter",
		"URL contains an invalid query parameter",
	)
}

func TestCategoriesCreateCategory(t *testing.T) {
	app := support.PrepareTestEnvironmentWithSignUpEmailWhitelist(t, support.TestUserEmail)
	signUpResp := support.MakeSignUpRequest(t, app, authenticationTransport.AuthenticationRequest{
		Email:    support.TestUserEmail,
		Password: support.TestUserPassword,
	})
	require.Equal(t, http.StatusCreated, signUpResp.Code)

	title := "Work & personal"
	resp := support.MakeJsonRequest(
		t,
		app,
		http.MethodPost,
		"/api/categories?title="+url.QueryEscape(title),
		nil,
		signUpResp.Result().Cookies()...,
	)
	require.Equal(t, http.StatusCreated, resp.Code)

	var respBody categoriesTransport.CategoryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, int64(1), respBody.Id, "unexpected category id")
	assert.Equal(t, title, respBody.Title, "unexpected category title")
}

const maximumCategoryTitleLength = 50
