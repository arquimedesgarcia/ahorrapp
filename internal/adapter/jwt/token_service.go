package jwt

import (
	"errors"
	"time"

	"ahorrapp/internal/domain/ports"

	jwtpkg "github.com/golang-jwt/jwt/v5"
)

type jwtService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string) ports.TokenService {
	if secret == "" {
		secret = "ahorrapp-dev-secret-change-me"
	}
	return &jwtService{secret: []byte(secret), ttl: 24 * time.Hour}
}

func (s *jwtService) Generate(userID, email string) (string, error) {
	claims := appClaims{
		Email: email,
		RegisteredClaims: jwtpkg.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwtpkg.NewNumericDate(time.Now()),
			ExpiresAt: jwtpkg.NewNumericDate(time.Now().Add(s.ttl)),
		},
	}
	token := jwtpkg.NewWithClaims(jwtpkg.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *jwtService) Validate(tokenStr string) (*ports.TokenClaims, error) {
	claims := &appClaims{}
	token, err := jwtpkg.ParseWithClaims(tokenStr, claims, func(t *jwtpkg.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtpkg.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	sub := claims.Subject
	if sub == "" {
		return nil, errors.New("missing subject")
	}

	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}

	return &ports.TokenClaims{
		UserID:    sub,
		Email:     claims.Email,
		ExpiresAt: exp,
	}, nil
}

// appClaims embeds RegisteredClaims and adds the user email.
type appClaims struct {
	jwtpkg.RegisteredClaims
	Email string `json:"email,omitempty"`
}
