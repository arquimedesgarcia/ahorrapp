package httpapi

import (
	"net/http"
	"strings"

	"ahorrapp/internal/domain/ports"
)

// JWTMiddleware validates the Authorization: Bearer <token> header and
// stores the user_id in the request context. If no token is present it
// leaves the context empty (handlers return 401). For backwards compat
// with existing dev/tests, the legacy X-User-ID header still works.
func JWTMiddleware(tokens ports.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prefer JWT bearer token
			if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				token := strings.TrimPrefix(auth, "Bearer ")
				claims, err := tokens.Validate(token)
				if err == nil && claims != nil && claims.UserID != "" {
					next.ServeHTTP(w, WithUserID(r, claims.UserID))
					return
				}
			}

			// Backwards-compat dev header (handlers fall back to this too)
			if fromHeader := r.Header.Get("X-User-ID"); fromHeader != "" {
				next.ServeHTTP(w, WithUserID(r, fromHeader))
				return
			}

			// No auth — let the handler decide (most will return 401).
			next.ServeHTTP(w, r)
		})
	}
}
