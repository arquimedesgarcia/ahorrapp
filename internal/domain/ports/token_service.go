package ports

import "time"

type TokenClaims struct {
	UserID    string
	Email     string
	ExpiresAt time.Time
}

type TokenService interface {
	Generate(userID, email string) (string, error)
	Validate(token string) (*TokenClaims, error)
}
