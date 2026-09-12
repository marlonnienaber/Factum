package transport

type CreateCategoryRequest struct {
	Title string `json:"title" binding:"min=1,max=50"`
	Color string `json:"color" binding:"required,len=7,lowercase,hexcolor"`
}
