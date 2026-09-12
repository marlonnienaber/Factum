package transport

type CategoryResponse struct {
	Id    int64  `db:"id" json:"id"`
	Title string `db:"title" json:"title"`
	Color string `db:"color" json:"color"`
}

type CreateCategoryResponse struct {
	Id int64 `db:"id" json:"id"`
}
