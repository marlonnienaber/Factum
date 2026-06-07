package transport

type SignUpRequest struct {
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=10,max=128"`
}
