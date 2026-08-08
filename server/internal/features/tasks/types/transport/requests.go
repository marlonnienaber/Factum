package transport

type CreateTaskRequest struct {
	Title       string  `json:"title" binding:"min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,min=1,max=2000"`
	Date        *string `json:"date" binding:"omitempty,datetime=2006-01-02"`
	CategoryId  int64   `json:"categoryId" binding:"gt=0"`
}
