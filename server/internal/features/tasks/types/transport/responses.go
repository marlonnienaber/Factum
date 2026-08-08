package transport

import "time"

type TaskResponse struct {
	Id          int64      `db:"id" json:"id"`
	CategoryId  int64      `db:"category_id" json:"categoryId"`
	Title       string     `db:"title" json:"title"`
	Description *string    `db:"description" json:"description,omitempty"`
	Date        *string    `db:"date" json:"date,omitempty"`
	Rank        string     `db:"rank" json:"rank"`
	Done        bool       `db:"done" json:"done"`
	TrashedDate *time.Time `db:"trashed_date" json:"trashedDate,omitempty"`
}
