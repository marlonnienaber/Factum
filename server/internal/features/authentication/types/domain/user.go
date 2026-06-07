package domain

import "github.com/marlonnienaber/Factum/server/internal/features/authentication/types/transport"

type User struct {
	Id           int64  `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

func (user *User) ToAuthenticationResponse() transport.AuthenticationResponse {
	return transport.AuthenticationResponse{
		Email: user.Email,
	}
}
