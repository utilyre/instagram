package middleware

import (
	"context"
	"net/http"

	"github.com/utilyre/instagram/internal/auth"
	"github.com/utilyre/xmate"
)

const prefixBearer = "Bearer "

var (
	ErrJWTMissing = xmate.NewHTTPError(http.StatusBadRequest, "missing or malformed jwt")
	ErrJWTInvalid = xmate.NewHTTPError(http.StatusUnauthorized, "invalid or expired jwt")
)

type ClaimsKey struct{}

func Authenticate(eh xmate.ErrorHandler, secret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return eh.HandleFunc(func(w http.ResponseWriter, r *http.Request) error {
			token := r.Header.Get("Authorization")
			if len(token) < len(prefixBearer)+1 {
				return ErrJWTMissing
			}

			claims, err := auth.Verify(secret, token[len(prefixBearer):])
			if err != nil {
				return ErrJWTInvalid
			}

			r2 := r.WithContext(context.WithValue(r.Context(), ClaimsKey{}, claims))
			next.ServeHTTP(w, r2)
			return nil
		})
	}
}
