package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"ahorrapp/internal/usecase"
)

func (h *ReceiptHandler) uploadReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "unable to read image file", http.StatusBadRequest)
		return
	}

	out, err := h.upload.Execute(r.Context(), usecase.UploadInput{UserID: userID, Data: data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(out)
}
