package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	healthHandler http.Handler,
	authHandler *AuthHandler,
	profileHandler *ProfileHandler,
	rankingHandler *RankingHandler,
	loyaltyHandler *LoyaltyHandler,
	registerReceiptRoutes func(chi.Router),
	jwtMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(v1 chi.Router) {
		// Public endpoints (no auth required)
		v1.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			healthHandler.ServeHTTP(w, req)
		})
		v1.Post("/auth/register", authHandler.Register)
		v1.Post("/auth/login", authHandler.Login)

		// Authenticated endpoints (JWT middleware required)
		v1.Group(func(authed chi.Router) {
			authed.Use(jwtMiddleware)

			authed.Get("/auth/me", authHandler.Me)
			authed.Get("/users/me/points", profileHandler.GetPoints)
			if loyaltyHandler != nil {
				authed.Get("/me/loyalty", loyaltyHandler.GetLoyalty)
			}
			authed.Get("/ranking/products/search", rankingHandler.Search)
			authed.Get("/products/{id}/prices", rankingHandler.ProductPrices)
			authed.Get("/search", rankingHandler.SearchV2)

			if registerReceiptRoutes != nil {
				registerReceiptRoutes(authed)
			}
		})
	})

	return r
}
