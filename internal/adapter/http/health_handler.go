package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"ahorrapp/internal/usecase"
)

type HealthQuery interface {
	Execute(ctx context.Context) usecase.HealthResponse
}

type HealthHandler struct {
	healthQuery HealthQuery
}

func NewHealthHandler(healthQuery HealthQuery) *HealthHandler {
	return &HealthHandler{healthQuery: healthQuery}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := h.healthQuery.Execute(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
