package httpapi

import (
	"net/http"
)

// RankingHandler is a stub for E8 — returns empty results until the
// price engine (E7) produces observations. This keeps the Flutter app
// from crashing while it consumes the published contract.
type RankingHandler struct{}

func NewRankingHandler() *RankingHandler { return &RankingHandler{} }

func (h *RankingHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}
	if q := r.URL.Query().Get("q"); q == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	// No price observations in the DB yet — return empty list per contract.
	writeJSON(w, http.StatusOK, map[string]any{"results": []any{}})
}
