package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(healthHandler http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			healthHandler.ServeHTTP(w, req)
		})
	})

	return r
}
