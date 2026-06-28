package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// newTestRouter builds a NewRouter with stub auth/profile/ranking handlers
// so existing receipt tests can keep calling a 2-arg helper.
func newTestRouter(healthHandler http.Handler, registerReceiptRoutes func(chi.Router)) http.Handler {
	return NewRouter(
		healthHandler,
		NewAuthHandler(nil),
		NewProfileHandler(nil),
		NewRankingHandler(),
		registerReceiptRoutes,
		JWTMiddleware(stubTokenService{}),
	)
}

var _ = http.StatusOK
