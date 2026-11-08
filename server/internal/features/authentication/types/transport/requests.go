package transport

type SignInRequest struct {
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"min=10,max=128"`
}

type SignUpRequest struct {
	Token    string `json:"token" binding:"emailVerificationToken"`
	Password string `json:"password" binding:"min=10,max=128"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"email"`
}
