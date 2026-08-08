package application

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/marlonnienaber/Factum/server/internal/configuration"
	authHandlers "github.com/marlonnienaber/Factum/server/internal/features/authentication/handlers"
	authServices "github.com/marlonnienaber/Factum/server/internal/features/authentication/services"
	"github.com/marlonnienaber/Factum/server/internal/features/authentication/stores"
	tasksHandlers "github.com/marlonnienaber/Factum/server/internal/features/tasks/handlers"
	tasksServices "github.com/marlonnienaber/Factum/server/internal/features/tasks/services"
	tasksStores "github.com/marlonnienaber/Factum/server/internal/features/tasks/stores"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"github.com/marlonnienaber/Factum/server/internal/validation"
)

func NewApplication(cfg configuration.Configuration, db *sqlx.DB) *gin.Engine {
	validation.RegisterJsonPropertyNames()

	userStore := stores.NewUserStore(db)
	authenticationService := authServices.NewAuthenticationService(userStore, cfg)
	taskStore := tasksStores.NewTaskStore(db)
	taskService := tasksServices.NewTaskService(db, taskStore)

	app := gin.Default()
	app.Use(shared.NewSessionConfigurationMiddleware(cfg, db))
	api := app.Group("/api")

	api.POST("/sign-up", authHandlers.NewSignUpHandler(authenticationService))
	api.POST("/sign-in", authHandlers.NewSignInHandler(authenticationService))
	api.POST("/sign-out", shared.RequireAuth(authHandlers.NewSignOutHandler(authenticationService)))
	api.GET("/current-authentication", shared.RequireAuth(authHandlers.NewGetCurrentAuthenticationHandler(userStore)))

	api.POST("/tasks", shared.RequireAuth(tasksHandlers.NewCreateTaskHandler(taskService)))
	api.GET("/tasks/undated", shared.RequireAuth(tasksHandlers.NewGetUndatedTasksHandler(taskStore)))
	api.GET("/tasks/dated-untrashed", shared.RequireAuth(tasksHandlers.NewGetDatedUntrashedTasksHandler(taskStore)))
	api.GET("/tasks/trashed", shared.RequireAuth(tasksHandlers.NewGetTrashedTasksHandler(taskStore)))

	return app
}
