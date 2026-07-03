package httpapi

import (
	"encoding/json"
	"net/http"

	"ahorrapp/internal/domain/entities"

	"github.com/go-chi/chi/v5"
)

func (h *ReceiptHandler) confirmReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	id := chi.URLParam(r, "id")
	var payload entities.ConfirmPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	result, err := h.confirm.Execute(r.Context(), id, userID, payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, "confirm receipt: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"points_earned": result.PointsEarned,
		"reasons":       result.Reasons,
	})
}
