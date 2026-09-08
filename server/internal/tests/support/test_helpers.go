package support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestUser1Email              = "ada@example.com"
	TestUser1Password           = "passwordPassword"
	TestUser2Email              = "grace@example.com"
	TestUser2Password           = "pAas5&WsHm}z#*]:W&qH"
	preparedPartitionTasksCount = 3
)

var defaultTestConfiguration = configuration.Configuration{
	Port:                ":3001",
	SessionCookieSecure: false,
	SessionSecretKey:    "test-session-secret-key",
}

func PrepareTestEnvironment(t *testing.T) *gin.Engine {
	t.Helper()
	return prepareTestEnvironmentWithConfig(t, defaultTestConfiguration)
}

func PrepareTestEnvironmentWithSignUpEmailWhitelist(t *testing.T, emails ...string) *gin.Engine {
	t.Helper()
	cfg := defaultTestConfiguration
	cfg.SignUpEmailWhitelist = emails
	return prepareTestEnvironmentWithConfig(t, cfg)
}

func prepareTestEnvironmentWithConfig(t *testing.T, cfg configuration.Configuration) *gin.Engine {
	t.Helper()
	databaseReset(t)
	return TestApp(cfg)
}

func MakeSignUpRequest(t *testing.T, app *gin.Engine, body authenticationTransport.AuthenticationRequest) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err, "serializing request failed")

	return MakeJsonRequest(t, app, http.MethodPost, "/api/sign-up", payload)
}

func MakeCreateCategoryRequest(
	t *testing.T,
	app *gin.Engine,
	title string,
	color string,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(categoriesTransport.CreateCategoryRequest{
		Title: title,
		Color: color,
	})
	require.NoError(t, err, "serializing request failed")

	return MakeJsonRequest(t, app, http.MethodPost, "/api/categories", payload, cookies...)
}

func MakeCreateTaskRequest(
	t *testing.T,
	app *gin.Engine,
	title string,
	description *string,
	date *shared.Date,
	categoryId int64,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(tasksTransport.CreateTaskRequest{
		Title:       title,
		Description: description,
		Date:        date,
		CategoryId:  categoryId,
	})
	require.NoError(t, err, "serializing request failed")

	return MakeJsonRequest(t, app, http.MethodPost, "/api/tasks", payload, cookies...)
}

func MakeTrashTaskRequest(
	t *testing.T,
	app *gin.Engine,
	taskId int64,
	trashedDate time.Time,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	path := fmt.Sprintf(
		"/api/tasks/%d/trash?%s",
		taskId,
		url.Values{"trashedDate": {trashedDate.Format(time.RFC3339Nano)}}.Encode(),
	)
	return MakeJsonRequest(t, app, http.MethodPatch, path, nil, cookies...)
}

func MakeToggleDoneRequest(
	t *testing.T,
	app *gin.Engine,
	taskId int64,
	rank *string,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	path := fmt.Sprintf("/api/tasks/%d/toggle-done", taskId)
	if rank != nil {
		path += "?" + url.Values{"rank": {*rank}}.Encode()
	}
	return MakeJsonRequest(t, app, http.MethodPatch, path, nil, cookies...)
}

func PrepareOpenUndatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	tasks := make([]tasksTransport.TaskResponse, 0, preparedPartitionTasksCount)
	for range preparedPartitionTasksCount {
		resp := MakeCreateTaskRequest(t, app, "Undated task", nil, nil, categoryId, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)

		var task tasksTransport.TaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&task), "failed to decode task response")
		task.CategoryId = categoryId
		task.Title = "Undated task"
		tasks = append(tasks, task)
	}
	return tasks
}

func PrepareOpenDatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	date, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")

	tasks := make([]tasksTransport.TaskResponse, 0, preparedPartitionTasksCount)
	for range preparedPartitionTasksCount {
		resp := MakeCreateTaskRequest(t, app, "Dated task", nil, &date, categoryId, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)

		var task tasksTransport.TaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&task), "failed to decode task response")
		task.CategoryId = categoryId
		task.Title = "Dated task"
		task.Date = &date
		tasks = append(tasks, task)
	}
	return tasks
}

func PrepareDoneDatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	tasks := PrepareOpenDatedUntrashedTasks(t, app, categoryId, cookies...)
	for index := range tasks {
		resp := MakeToggleDoneRequest(t, app, tasks[index].Id, tasks[index].Rank, cookies...)
		require.Equal(t, http.StatusNoContent, resp.Code)
		tasks[index].Done = true
	}
	return tasks
}

func PrepareTrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.TaskResponse {
	t.Helper()

	date, err := shared.ParseDate("2026-03-25")
	require.NoError(t, err, "parsing test date failed")
	trashedDate := time.Date(2026, time.March, 26, 10, 30, 0, 0, time.UTC)

	trashTask := func(task tasksTransport.TaskResponse) tasksTransport.TaskResponse {
		resp := MakeTrashTaskRequest(t, app, task.Id, trashedDate, cookies...)
		require.Equal(t, http.StatusNoContent, resp.Code)
		task.Rank = nil
		task.TrashedDate = &trashedDate
		return task
	}

	openUndatedResp := MakeCreateTaskRequest(t, app, "Open undated task", nil, nil, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, openUndatedResp.Code)
	var openUndatedTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(openUndatedResp.Body).Decode(&openUndatedTask),
		"failed to decode task response",
	)
	openUndatedTask.CategoryId = categoryId
	openUndatedTask.Title = "Open undated task"

	openDatedResp := MakeCreateTaskRequest(t, app, "Open dated task", nil, &date, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, openDatedResp.Code)
	var openDatedTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(openDatedResp.Body).Decode(&openDatedTask),
		"failed to decode task response",
	)
	openDatedTask.CategoryId = categoryId
	openDatedTask.Title = "Open dated task"
	openDatedTask.Date = &date

	doneDatedResp := MakeCreateTaskRequest(t, app, "Done dated task", nil, &date, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, doneDatedResp.Code)
	var doneDatedTask tasksTransport.TaskResponse
	require.NoError(
		t,
		json.NewDecoder(doneDatedResp.Body).Decode(&doneDatedTask),
		"failed to decode task response",
	)
	doneDatedTask.CategoryId = categoryId
	doneDatedTask.Title = "Done dated task"
	doneDatedTask.Date = &date
	require.NotNil(t, doneDatedTask.Rank, "expected created task to have a rank")
	toggleDoneResp := MakeToggleDoneRequest(t, app, doneDatedTask.Id, doneDatedTask.Rank, cookies...)
	require.Equal(t, http.StatusNoContent, toggleDoneResp.Code)
	doneDatedTask.Done = true

	return []tasksTransport.TaskResponse{
		trashTask(doneDatedTask),
		trashTask(openDatedTask),
		trashTask(openUndatedTask),
	}
}

func MakeJsonRequest(
	t *testing.T,
	app *gin.Engine,
	method string,
	path string,
	payload []byte,
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	return resp
}

func AssertValidationErrorResponse(
	t *testing.T,
	resp *httptest.ResponseRecorder,
	expectedField string,
	expectedViolationDescription string,
) {
	t.Helper()

	require.Equal(t, http.StatusBadRequest, resp.Code)

	var respBody shared.DetailedErrorResponse[validation.ValidationErrorDetails]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	require.Equal(t, "validation_error", respBody.ErrorCode, "unexpected errorCode")
	require.Equal(t, "Request body did not match the expected form.", respBody.Message, "unexpected message")
	require.Len(t, respBody.Details.Violations, 1, "expected exactly one violation")

	violation := respBody.Details.Violations[0]
	assert.Equal(t, expectedField, violation.Field, "unexpected field violated")
	assert.Equal(t, expectedViolationDescription, violation.Description, "unexpected violation description")
}

func AssertErrorResponse(
	t *testing.T,
	resp *httptest.ResponseRecorder,
	expectedStatusCode int,
	expectedErrorCode string,
	expectedMessage string,
) {
	t.Helper()

	require.Equal(t, expectedStatusCode, resp.Code)

	var respBody shared.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody), "failed to decode response")

	assert.Equal(t, expectedErrorCode, respBody.ErrorCode, "unexpected error code")
	assert.Equal(t, expectedMessage, respBody.Message, "unexpected message")
}
