package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ahorrapp/internal/domain/entities"

	"github.com/go-chi/chi/v5"
)

func (h *ReceiptHandler) confirmReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	var payload entities.ConfirmPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	result, err := h.confirm.Execute(r.Context(), id, userID, payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("confirm receipt: %v", err), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"points_earned": result.PointsEarned,
		"reasons":       result.Reasons,
	})
}
