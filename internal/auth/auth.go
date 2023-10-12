package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims

	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func Generate(secret []byte, claims *Claims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(72 * time.Hour))
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func Verify(secret []byte, token string) (*Claims, error) {
	claims := new(Claims)

	if _, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) { return secret, nil },
	); err != nil {
		return nil, err
	}

	return claims, nil
}
