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
	authenticationTransport "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"
	categoriesTransport "github.com/marlonnienaber/Factum/server/internal/features/categories/types/transport"
	tasksTransport "github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/injections"
	"github.com/marlonnienaber/Factum/server/internal/shared"
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

var TestTaskDate = shared.NewDate(time.Date(2026, time.March, 25, 0, 0, 0, 0, time.UTC))

var defaultTestEnvironment = injections.Environment{
	ServerPort:          "3001",
	ClientPort:          "5173",
	AppUrl:              "http://localhost:5173",
	SessionCookieSecure: false,
	SessionSecretKey:    "test-session-secret-key",
}

func PrepareTestEnvironment(t *testing.T, signUpEmailWhitelist ...string) SystemUnderTest {
	t.Helper()
	env := defaultTestEnvironment
	env.SignUpEmailWhitelist = signUpEmailWhitelist
	databaseReset(t)
	return NewSystemUnderTest(env)
}

func MakeVerifyEmailRequest(
	t *testing.T,
	sut SystemUnderTest,
	email string,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(authenticationTransport.VerifyEmailRequest{
		Email: email,
	})
	require.NoError(t, err, "serializing verify-email request failed")
	return MakeJsonRequest(t, sut.App, http.MethodPost, "/api/verify-email", payload)
}

func MakeVerifyEmailAndSignUpRequests(
	t *testing.T,
	sut SystemUnderTest,
	email string,
	password string,
) *httptest.ResponseRecorder {
	t.Helper()

	sentEmailCount := len(sut.EmailService.SentEmails)
	verifyEmailResponse := MakeVerifyEmailRequest(t, sut, email)
	require.Equal(t, http.StatusNoContent, verifyEmailResponse.Code)
	require.Len(t, sut.EmailService.SentEmails, sentEmailCount+1, "expected an email verification link to be sent")

	sentEmail := sut.EmailService.SentEmails[len(sut.EmailService.SentEmails)-1]
	verificationLink, err := url.Parse(sentEmail.VerificationLink)
	require.NoError(t, err, "parsing email verification link failed")
	verificationToken := verificationLink.Query().Get("token")
	require.NotEmpty(t, verificationToken, "email verification link did not contain a token")

	signUpPayload, err := json.Marshal(authenticationTransport.SignUpRequest{
		Token:    verificationToken,
		Password: password,
	})
	require.NoError(t, err, "serializing sign-up request failed")
	signUpResponse := MakeJsonRequest(t, sut.App, http.MethodPost, "/api/sign-up", signUpPayload)
	require.Equal(t, http.StatusCreated, signUpResponse.Code)

	return signUpResponse
}

func MakeVerifyEmailSignUpAndSignInRequests(
	t *testing.T,
	sut SystemUnderTest,
	email string,
	password string,
) *httptest.ResponseRecorder {
	t.Helper()

	MakeVerifyEmailAndSignUpRequests(t, sut, email, password)

	signInPayload, err := json.Marshal(authenticationTransport.SignInRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err, "serializing sign-in request failed")
	signInResponse := MakeJsonRequest(t, sut.App, http.MethodPost, "/api/sign-in", signInPayload)
	require.Equal(t, http.StatusOK, signInResponse.Code)

	return signInResponse
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
	cookies ...*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()

	path := fmt.Sprintf("/api/tasks/%d/toggle-done", taskId)
	return MakeJsonRequest(t, app, http.MethodPatch, path, nil, cookies...)
}

func PrepareOpenUndatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	cookies ...*http.Cookie,
) []tasksTransport.UndatedTaskResponse {
	t.Helper()

	tasks := make([]tasksTransport.UndatedTaskResponse, 0, preparedPartitionTasksCount)
	for range preparedPartitionTasksCount {
		resp := MakeCreateTaskRequest(t, app, "Undated task", nil, nil, categoryId, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)

		var response tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode task response")
		tasks = append(tasks, tasksTransport.UndatedTaskResponse{
			Id:         response.Id,
			CategoryId: categoryId,
			Title:      "Undated task",
			Rank:       response.Rank,
		})
	}
	return tasks
}

func PrepareOpenDatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	date shared.Date,
	cookies ...*http.Cookie,
) []tasksTransport.DatedUntrashedTaskResponse {
	t.Helper()

	tasks := make([]tasksTransport.DatedUntrashedTaskResponse, 0, preparedPartitionTasksCount)
	for range preparedPartitionTasksCount {
		resp := MakeCreateTaskRequest(t, app, "Dated task", nil, &date, categoryId, cookies...)
		require.Equal(t, http.StatusCreated, resp.Code)

		var response tasksTransport.CreateTaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response), "failed to decode task response")
		tasks = append(tasks, tasksTransport.DatedUntrashedTaskResponse{
			Id:         response.Id,
			CategoryId: categoryId,
			Title:      "Dated task",
			Date:       date,
			Rank:       response.Rank,
		})
	}
	return tasks
}

func PrepareDoneDatedUntrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	date shared.Date,
	cookies ...*http.Cookie,
) []tasksTransport.DatedUntrashedTaskResponse {
	t.Helper()

	tasks := PrepareOpenDatedUntrashedTasks(t, app, categoryId, date, cookies...)
	for index := range tasks {
		resp := MakeToggleDoneRequest(t, app, tasks[index].Id, cookies...)
		require.Equal(t, http.StatusOK, resp.Code)
		var toggleResponse tasksTransport.ToggleTaskDoneResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&toggleResponse), "failed to decode toggle task done response")
		require.NotNil(t, toggleResponse.Rank, "expected toggled task to have a rank")
		tasks[index].Date = toggleResponse.Date
		tasks[index].Rank = *toggleResponse.Rank
	}
	return tasks
}

func PrepareTrashedTasks(
	t *testing.T,
	app *gin.Engine,
	categoryId int64,
	date shared.Date,
	cookies ...*http.Cookie,
) []tasksTransport.TrashedTaskResponse {
	t.Helper()

	trashedDate := time.Date(2026, time.March, 26, 10, 30, 0, 0, time.UTC)

	trashTask := func(task tasksTransport.TrashedTaskResponse) tasksTransport.TrashedTaskResponse {
		resp := MakeTrashTaskRequest(t, app, task.Id, trashedDate, cookies...)
		require.Equal(t, http.StatusOK, resp.Code)
		var trashTaskResponse tasksTransport.TrashTaskResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&trashTaskResponse), "failed to decode trash task response")
		task.Rank = trashTaskResponse.Rank
		return task
	}

	openUndatedResp := MakeCreateTaskRequest(t, app, "Open undated task", nil, nil, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, openUndatedResp.Code)
	var openUndatedCreateResponse tasksTransport.CreateTaskResponse
	require.NoError(
		t,
		json.NewDecoder(openUndatedResp.Body).Decode(&openUndatedCreateResponse),
		"failed to decode task response",
	)
	openUndatedTask := tasksTransport.TrashedTaskResponse{
		Id:         openUndatedCreateResponse.Id,
		CategoryId: categoryId,
		Title:      "Open undated task",
	}

	openDatedResp := MakeCreateTaskRequest(t, app, "Open dated task", nil, &date, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, openDatedResp.Code)
	var openDatedCreateResponse tasksTransport.CreateTaskResponse
	require.NoError(
		t,
		json.NewDecoder(openDatedResp.Body).Decode(&openDatedCreateResponse),
		"failed to decode task response",
	)
	openDatedTask := tasksTransport.TrashedTaskResponse{
		Id:         openDatedCreateResponse.Id,
		CategoryId: categoryId,
		Title:      "Open dated task",
		Date:       &date,
	}

	doneDatedResp := MakeCreateTaskRequest(t, app, "Done dated task", nil, &date, categoryId, cookies...)
	require.Equal(t, http.StatusCreated, doneDatedResp.Code)
	var doneDatedCreateResponse tasksTransport.CreateTaskResponse
	require.NoError(
		t,
		json.NewDecoder(doneDatedResp.Body).Decode(&doneDatedCreateResponse),
		"failed to decode task response",
	)
	doneDatedTask := tasksTransport.TrashedTaskResponse{
		Id:         doneDatedCreateResponse.Id,
		CategoryId: categoryId,
		Title:      "Done dated task",
		Date:       &date,
	}
	toggleDoneResp := MakeToggleDoneRequest(t, app, doneDatedTask.Id, cookies...)
	require.Equal(t, http.StatusOK, toggleDoneResp.Code)
	doneDatedTask.Done = true

	doneDatedTask = trashTask(doneDatedTask)
	openDatedTask = trashTask(openDatedTask)
	openUndatedTask = trashTask(openUndatedTask)

	return []tasksTransport.TrashedTaskResponse{
		openUndatedTask,
		openDatedTask,
		doneDatedTask,
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

	var respBody shared.DetailedErrorResponse[shared.ValidationErrorDetails]
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
