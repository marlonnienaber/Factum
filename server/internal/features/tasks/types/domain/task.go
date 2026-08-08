package domain

import (
	"time"

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
