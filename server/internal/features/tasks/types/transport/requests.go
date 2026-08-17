package transport

import "github.com/marlonnienaber/Factum/server/internal/shared"

type CreateTaskRequest struct {
	Title       string       `json:"title" binding:"min=1,max=100"`
	Description *string      `json:"description" binding:"omitempty,min=1,max=500"`
	Date        *shared.Date `json:"date"`
	CategoryId  int64        `json:"categoryId" binding:"gt=0"`
}

type UpdateTaskRequest struct {
	Id          int64        `json:"id" binding:"gt=0"`
	Title       string       `json:"title" binding:"min=1,max=100"`
	Description *string      `json:"description" binding:"omitempty,min=1,max=500"`
	Date        *shared.Date `json:"date"`
	CategoryId  int64        `json:"categoryId" binding:"gt=0"`
}

type ChangeTaskDateRequest struct {
	Id   int64        `json:"id" binding:"gt=0"`
	Date *shared.Date `json:"date"`
	Rank string       `json:"rank" binding:"rank"`
}
