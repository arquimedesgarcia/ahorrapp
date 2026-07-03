package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"ahorrapp/internal/domain/ports"
	"ahorrapp/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RankingHandler struct {
	uc *usecase.RankingUseCase
}

func NewRankingHandler(uc *usecase.RankingUseCase) *RankingHandler {
	return &RankingHandler{uc: uc}
}

func (h *RankingHandler) ProductPrices(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	productID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(productID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	opts := ports.RankingQueryOptions{}
	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			opts.Lat = &lat
		}
	}
	if longStr := r.URL.Query().Get("long"); longStr != "" {
		if long, err := strconv.ParseFloat(longStr, 64); err == nil {
			opts.Long = &long
		}
	}
	if radiusStr := r.URL.Query().Get("radius_km"); radiusStr != "" {
		if radius, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			opts.RadiusKm = &radius
		}
	}

	resp, err := h.uc.GetProductRanking(r.Context(), productID, opts)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *RankingHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	q := r.URL.Query().Get("q")
	if strings.TrimSpace(q) == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	resp, err := h.uc.SearchProductsLegacy(r.Context(), q)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *RankingHandler) SearchV2(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	q := r.URL.Query().Get("q")
	if strings.TrimSpace(q) == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	if len(strings.TrimSpace(q)) < 3 {
		writeError(w, http.StatusBadRequest, "query must be at least 3 characters")
		return
	}

	resp, err := h.uc.SearchProducts(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
