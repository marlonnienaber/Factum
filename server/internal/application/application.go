package application

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	authHandlers "github.com/marlonnienaber/Factum/server/internal/features/authentication/handlers"
	authServices "github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	categoriesHandlers "github.com/marlonnienaber/Factum/server/internal/features/categories/handlers"
	categoriesServices "github.com/marlonnienaber/Factum/server/internal/features/categories/services"
	categoriesStores "github.com/marlonnienaber/Factum/server/internal/features/categories/stores"
	tasksHandlers "github.com/marlonnienaber/Factum/server/internal/features/tasks/handlers"
	tasksServices "github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	tasksStores "github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewApplication(cfg configuration.Configuration, db *sqlx.DB, injectionOptions ...InjectionOption) *gin.Engine {
	gin.EnableJsonDecoderDisallowUnknownFields()
	validation.RegisterJsonPropertyNames()
	validation.RegisterRankValidation()
	validation.RegisterEmailVerificationTokenValidation()

	injectableComponents := injectableAppComponents{
		emailService: authServices.NewDevelopmentEmailService(),
		clock:        shared.NewDefaultClock(),
	}
	for _, option := range injectionOptions {
		option.inject(&injectableComponents)
	}

	userStore := stores.NewUserStore(db)
	emailVerificationTokenRecordStore := stores.NewEmailVerificationTokenRecordStore(db)
	authenticationService := authServices.NewAuthenticationService(
		db,
		userStore,
		emailVerificationTokenRecordStore,
		injectableComponents.emailService,
		injectableComponents.clock,
		cfg,
	)
	categoryStore := categoriesStores.NewCategoryStore(db)
	categoryService := categoriesServices.NewCategoryService(db, categoryStore)
	taskStore := tasksStores.NewTaskStore(db)
	taskService := tasksServices.NewTaskService(db, taskStore, categoryStore)

	app := gin.Default()
	app.Use(shared.NewSessionConfigurationMiddleware(cfg, db))
	api := app.Group("/api")

	api.POST(
		"/verify-email",
		authHandlers.NewVerifyEmailHandler(authenticationService),
	)
	api.POST(
		"/sign-up",
		authHandlers.NewSignUpHandler(authenticationService),
	)
	api.POST(
		"/sign-in",
		authHandlers.NewSignInHandler(authenticationService),
	)
	api.POST(
		"/sign-out",
		shared.RequireAuth(authHandlers.NewSignOutHandler(authenticationService)),
	)
	api.GET(
		"/current-authentication",
		shared.RequireAuth(authHandlers.NewGetCurrentAuthenticationHandler(userStore)),
	)

	api.GET(
		"/categories",
		shared.RequireAuth(categoriesHandlers.NewGetCategoriesHandler(categoryStore)),
	)
	api.POST(
		"/categories",
		shared.RequireAuth(categoriesHandlers.NewCreateCategoryHandler(categoryService)),
	)
	api.DELETE(
		"/categories/:id",
		shared.RequireAuth(categoriesHandlers.NewDeleteCategoryHandler(categoryStore)),
	)

	api.POST(
		"/tasks",
		shared.RequireAuth(tasksHandlers.NewCreateTaskHandler(taskService)),
	)
	api.PUT(
		"/tasks",
		shared.RequireAuth(tasksHandlers.NewUpdateTaskHandler(taskService)),
	)
	api.PATCH(
		"/tasks/:id/drag-and-drop",
		shared.RequireAuth(tasksHandlers.NewTaskDragAndDropHandler(taskService)),
	)
	api.PATCH(
		"/tasks/batch-complete",
		shared.RequireAuth(tasksHandlers.NewBatchCompleteTasksHandler(taskService)),
	)
	api.PATCH(
		"/tasks/batch-open",
		shared.RequireAuth(tasksHandlers.NewBatchOpenTasksHandler(taskService)),
	)
	api.PATCH(
		"/tasks/batch-trash",
		shared.RequireAuth(tasksHandlers.NewBatchTrashTasksHandler(taskService)),
	)
	api.PATCH(
		"/tasks/:id/trash",
		shared.RequireAuth(tasksHandlers.NewTrashTaskHandler(taskService)),
	)
	api.PATCH(
		"/tasks/:id/restore",
		shared.RequireAuth(tasksHandlers.NewRestoreTaskHandler(taskService)),
	)
	api.PATCH(
		"/tasks/:id/toggle-done",
		shared.RequireAuth(tasksHandlers.NewToggleTaskDoneHandler(taskService)),
	)
	api.DELETE(
		"/tasks/trashed",
		shared.RequireAuth(tasksHandlers.NewDeleteTrashedTasksHandler(taskStore)),
	)
	api.GET(
		"/tasks/trashed/exists",
		shared.RequireAuth(tasksHandlers.NewGetTrashedTasksExistHandler(taskStore)),
	)
	api.GET(
		"/tasks/today-progress",
		shared.RequireAuth(tasksHandlers.NewGetTodayProgressHandler(taskStore)),
	)
	api.GET(
		"/tasks/undated",
		shared.RequireAuth(tasksHandlers.NewGetUndatedTasksHandler(taskStore)),
	)
	api.GET(
		"/tasks/dated-untrashed-done",
		shared.RequireAuth(tasksHandlers.NewGetDatedUntrashedDoneTasksHandler(taskStore)),
	)
	api.GET(
		"/tasks/dated-untrashed-done-today",
		shared.RequireAuth(tasksHandlers.NewGetDatedUntrashedDoneTodayTasksHandler(taskStore)),
	)
	api.GET(
		"/tasks/dated-untrashed-open-today",
		shared.RequireAuth(tasksHandlers.NewGetDatedUntrashedOpenTodayTasksHandler(taskStore)),
	)
	api.GET(
		"/tasks/dated-untrashed-open-non-today",
		shared.RequireAuth(tasksHandlers.NewGetDatedUntrashedOpenNonTodayTasksHandler(taskStore)),
	)
	api.GET(
		"/tasks/trashed",
		shared.RequireAuth(tasksHandlers.NewGetTrashedTasksHandler(taskStore)),
	)

	return app
}
