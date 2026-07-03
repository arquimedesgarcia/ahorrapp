package httpapi

import (
	"net/http"

	"ahorrapp/internal/usecase"
)

type LoyaltyHandler struct {
	query *usecase.LoyaltyQueryUseCase
}

func NewLoyaltyHandler(query *usecase.LoyaltyQueryUseCase) *LoyaltyHandler {
	return &LoyaltyHandler{query: query}
}

func (h *LoyaltyHandler) GetLoyalty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	resp, err := h.query.GetLoyalty(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
