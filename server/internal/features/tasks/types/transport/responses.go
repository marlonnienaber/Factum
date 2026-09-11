package transport

import (
	"time"

	"github.com/marlonnienaber/Factum/server/internal/shared"
)

type TaskResponse struct {
	Id          int64        `db:"id" json:"id"`
	CategoryId  int64        `db:"category_id" json:"categoryId"`
	Title       string       `db:"title" json:"title"`
	Description *string      `db:"description" json:"description,omitempty"`
	Date        *shared.Date `db:"date" json:"date,omitempty"`
	Rank        *string      `db:"rank" json:"rank,omitempty"`
	Done        bool         `db:"done" json:"done"`
	TrashedDate *time.Time   `db:"trashed_date" json:"trashedDate,omitempty"`
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

type BatchCompleteResponse struct {
	DatedUntrashedDoneTasks []TaskResponse `json:"datedUntrashedDoneTasks"`
	TrashedTasks            []TaskResponse `json:"trashedTasks"`
}

type BatchOpenResponse struct {
	DatedUntrashedOpenTasks []TaskResponse `json:"datedUntrashedOpenTasks"`
	TrashedTasks            []TaskResponse `json:"trashedTasks"`
}
