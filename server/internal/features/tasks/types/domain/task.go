package domain

import (
	"time"

	"github.com/marlonnienaber/Factum/server/internal/features/tasks/types/transport"
	"github.com/marlonnienaber/Factum/server/internal/shared"
)

type Task struct {
	Id          int64        `db:"id"`
	UserId      int64        `db:"user_id"`
	CategoryId  int64        `db:"category_id"`
	Title       string       `db:"title"`
	Description *string      `db:"description"`
	Date        *shared.Date `db:"date"`
	Rank        *string      `db:"rank"`
	Done        bool         `db:"done"`
	TrashedDate *time.Time   `db:"trashed_date"`
}

func (task *Task) ToTaskResponse() *transport.TaskResponse {
	return &transport.TaskResponse{
		Id:          task.Id,
		CategoryId:  task.CategoryId,
		Title:       task.Title,
		Description: task.Description,
		Date:        task.Date,
		Rank:        task.Rank,
		Done:        task.Done,
		TrashedDate: task.TrashedDate,
	}
}
