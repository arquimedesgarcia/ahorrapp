package httpapi

import (
	"errors"
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
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil || lat < -90 || lat > 90 {
			writeError(w, http.StatusBadRequest, "invalid lat parameter (expected -90..90)")
			return
		}
		opts.Lat = &lat
	}
	if longStr := r.URL.Query().Get("long"); longStr != "" {
		long, err := strconv.ParseFloat(longStr, 64)
		if err != nil || long < -180 || long > 180 {
			writeError(w, http.StatusBadRequest, "invalid long parameter (expected -180..180)")
			return
		}
		opts.Long = &long
	}
	if radiusStr := r.URL.Query().Get("radius_km"); radiusStr != "" {
		radius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil || radius <= 0 {
			writeError(w, http.StatusBadRequest, "invalid radius_km parameter (expected > 0)")
			return
		}
		opts.RadiusKm = &radius
	}
	// lat and long must be provided together; partial proximity is
	// rejected so the ranking never silently drops one coordinate.
	if (opts.Lat == nil) != (opts.Long == nil) {
		writeError(w, http.StatusBadRequest, "lat and long must be provided together")
		return
	}

	resp, err := h.uc.GetProductRanking(r.Context(), productID, opts)
	if err != nil {
		if errors.Is(err, ports.ErrProductNotFound) {
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
