package httpapi

import (
	"net/http"

	"ahorrapp/internal/usecase"
)

type ProfileHandler struct {
	profile *usecase.ProfileUseCase
}

func NewProfileHandler(profile *usecase.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{profile: profile}
}

func (h *ProfileHandler) GetPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	resp, err := h.profile.GetPoints(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
