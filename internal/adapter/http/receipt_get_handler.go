package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *ReceiptHandler) getReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	out, err := h.get.Execute(r.Context(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if out == nil {
		http.Error(w, "receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}
