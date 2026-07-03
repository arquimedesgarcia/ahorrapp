package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
	"ahorrapp/internal/usecase"
)

type stubRankingRepoForHandler struct {
	ranking    []entities.PriceAggregate
	search     []entities.ProductSearchResult
	productMap map[string]string
}

func (s *stubRankingRepoForHandler) RecomputeAggregate(context.Context, string, string, string, int) error {
	return nil
}
func (s *stubRankingRepoForHandler) RecomputeAll(context.Context, int) error { return nil }
func (s *stubRankingRepoForHandler) GetProductRanking(ctx context.Context, productID string, opts ports.RankingQueryOptions) ([]entities.PriceAggregate, error) {
	return s.ranking, nil
}
func (s *stubRankingRepoForHandler) SearchProducts(ctx context.Context, query string) ([]entities.ProductSearchResult, error) {
	return s.search, nil
}
func (s *stubRankingRepoForHandler) GetProductName(ctx context.Context, productID string) (string, error) {
	if name, ok := s.productMap[productID]; ok {
		return name, nil
	}
	return "Unknown", nil
}

func newRankingTestHandler(repo ports.RankingRepository) http.Handler {
	uc := usecase.NewRankingUseCase(repo)
	rankingHandler := NewRankingHandler(uc)
	return NewRouter(
		NewHealthHandler(fakeHealthUseCase{}),
		NewAuthHandler(nil),
		NewProfileHandler(nil),
		rankingHandler,
		nil,
		nil,
		JWTMiddleware(stubTokenService{}),
	)
}

func TestRankingHandler_ProductPrices_Returns200WithRanking(t *testing.T) {
	repo := &stubRankingRepoForHandler{
		ranking: []entities.PriceAggregate{
			{StoreID: "s1", StoreName: "Store A", Currency: "USD", AveragePrice: 1.00, MinPrice: 0.90, SampleCount: 5},
			{StoreID: "s2", StoreName: "Store B", Currency: "USD", AveragePrice: 1.50, MinPrice: 1.40, SampleCount: 3},
		},
		productMap: map[string]string{"00000000-0000-0000-0000-000000000001": "Test Product"},
	}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/00000000-0000-0000-0000-000000000001/prices", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["currency_rankings"]; !ok {
		t.Error("expected currency_rankings in response")
	}
}

func TestRankingHandler_ProductPrices_400ForInvalidUUID(t *testing.T) {
	repo := &stubRankingRepoForHandler{}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/not-a-uuid/prices", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRankingHandler_ProductPrices_401ForMissingAuth(t *testing.T) {
	repo := &stubRankingRepoForHandler{}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/00000000-0000-0000-0000-000000000001/prices", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRankingHandler_Search_Returns200WithResults(t *testing.T) {
	repo := &stubRankingRepoForHandler{
		search: []entities.ProductSearchResult{
			{ProductID: "p1", ProductName: "Arroz", BestPrices: map[string]*entities.PriceAggregate{
				"USD": {StoreName: "Store A", Currency: "USD", AveragePrice: 1.25},
			}},
		},
	}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ranking/products/search?q=arroz", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	results, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("expected results array")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRankingHandler_Search_400ForMissingQuery(t *testing.T) {
	repo := &stubRankingRepoForHandler{}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ranking/products/search", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRankingHandler_Search_401ForMissingAuth(t *testing.T) {
	repo := &stubRankingRepoForHandler{}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ranking/products/search?q=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRankingHandler_SearchV2_Returns200WithResults(t *testing.T) {
	repo := &stubRankingRepoForHandler{
		search: []entities.ProductSearchResult{
			{ProductID: "p1", ProductName: "Arroz", BestPrices: map[string]*entities.PriceAggregate{
				"USD": {StoreName: "Store A", Currency: "USD", AveragePrice: 1.25},
			}},
		},
	}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=arroz", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["results"]; !ok {
		t.Error("expected results in response")
	}
}

func TestRankingHandler_SearchV2_400ForMissingQuery(t *testing.T) {
	repo := &stubRankingRepoForHandler{}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRankingHandler_SearchV2_400ForShortQuery(t *testing.T) {
	repo := &stubRankingRepoForHandler{}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=ab", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRankingHandler_LegacySearch_ReturnsFlatStoresArray(t *testing.T) {
	repo := &stubRankingRepoForHandler{
		search: []entities.ProductSearchResult{
			{ProductID: "p1", ProductName: "Arroz", BestPrices: map[string]*entities.PriceAggregate{
				"USD": {StoreID: "s1", StoreName: "Central Market", Currency: "USD", AveragePrice: 1.25, SampleCount: 45},
			}},
		},
	}
	router := newRankingTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ranking/products/search?q=arroz", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	results := resp["results"].([]interface{})
	r0 := results[0].(map[string]interface{})
	if _, ok := r0["stores"]; !ok {
		t.Error("expected flat 'stores' array in legacy response")
	}
	stores := r0["stores"].([]interface{})
	if len(stores) != 1 {
		t.Errorf("expected 1 store, got %d", len(stores))
	}
}
