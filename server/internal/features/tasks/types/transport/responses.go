package transport

import (
	"time"

	"github.com/marlonnienaber/Factum/server/internal/shared"
)

type UpdateTaskResponse struct {
	Id          int64        `db:"id" json:"id"`
	CategoryId  int64        `db:"category_id" json:"categoryId"`
	Title       string       `db:"title" json:"title"`
	Description *string      `db:"description" json:"description,omitempty"`
	Date        *shared.Date `db:"date" json:"date,omitempty"`
	Rank        *string      `db:"rank" json:"rank,omitempty"`
	Done        bool         `db:"done" json:"done"`
	TrashedDate *time.Time   `db:"trashed_date" json:"trashedDate,omitempty"`
}

type DatedUntrashedTaskResponse struct {
	Id          int64       `db:"id" json:"id"`
	CategoryId  int64       `db:"category_id" json:"categoryId"`
	Title       string      `db:"title" json:"title"`
	Description *string     `db:"description" json:"description,omitempty"`
	Date        shared.Date `db:"date" json:"date"`
	Rank        string      `db:"rank" json:"rank"`
}

type GetDatedUntrashedTasksResponse struct {
	Tasks   []DatedUntrashedTaskResponse `json:"tasks"`
	HasMore bool                         `json:"hasMore"`
}

type CreateTaskResponse struct {
	Id   int64  `db:"id" json:"id"`
	Rank string `db:"rank" json:"rank"`
}

type TaskDragAndDropResponse struct {
	Rank string `json:"rank"`
}

type RestoreTaskResponse struct {
	Rank string `json:"rank"`
}

type ToggleTaskDoneResponse struct {
	Date shared.Date `json:"date"`
	Rank *string     `json:"rank,omitempty"`
}

type UndatedTaskResponse struct {
	Id          int64   `db:"id" json:"id"`
	CategoryId  int64   `db:"category_id" json:"categoryId"`
	Title       string  `db:"title" json:"title"`
	Description *string `db:"description" json:"description,omitempty"`
	Rank        string  `db:"rank" json:"rank"`
}

type GetUndatedTasksResponse struct {
	Tasks   []UndatedTaskResponse `json:"tasks"`
	HasMore bool                  `json:"hasMore"`
}

type TrashedTaskResponse struct {
	Id          int64        `db:"id" json:"id"`
	CategoryId  int64        `db:"category_id" json:"categoryId"`
	Title       string       `db:"title" json:"title"`
	Description *string      `db:"description" json:"description,omitempty"`
	Date        *shared.Date `db:"date" json:"date,omitempty"`
	Done        bool         `db:"done" json:"done"`
	TrashedDate time.Time    `db:"trashed_date" json:"trashedDate"`
}

type GetTrashedTasksResponse struct {
	Tasks   []TrashedTaskResponse `json:"tasks"`
	HasMore bool                  `json:"hasMore"`
}

type BatchCompleteTaskResponse struct {
	Id   int64       `json:"id"`
	Date shared.Date `json:"date"`
	Rank string      `json:"rank"`
}

type BatchOpenTaskResponse struct {
	Id   int64  `json:"id"`
	Rank string `json:"rank"`
}
