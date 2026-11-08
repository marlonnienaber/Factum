package transport

type AuthenticationRequest struct {
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=10,max=128"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"email"`
}
