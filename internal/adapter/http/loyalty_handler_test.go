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

// stubConnloiledLoyaltyRepo isolates the LoyaltyHandler tests from any real
// database. It records the userID observed via the injected query use
// case so the cross-user test can assert that only the JWT user's data is
// requested.
type stubLoyaltyRepo struct {
	balance    int
	history    []entities.LoyaltyTransaction
	contrib    ports.ContributorStats
	lastUserID string
	balanceErr error
	historyErr error
}

func (s *stubLoyaltyRepo) AwardForReceipt(context.Context, string, string, int, string) error {
	return nil
}
func (s *stubLoyaltyRepo) DailyGrantCount(context.Context, string) (int, error) {
	return 0, nil
}
func (s *stubLoyaltyRepo) Balance(_ context.Context, userID string) (int, error) {
	s.lastUserID = userID
	return s.balance, s.balanceErr
}
func (s *stubLoyaltyRepo) History(_ context.Context, userID string, _ int) ([]entities.LoyaltyTransaction, error) {
	s.lastUserID = userID
	return s.history, s.historyErr
}
func (s *stubLoyaltyRepo) ContributorStats(_ context.Context, userID string) (ports.ContributorStats, error) {
	s.lastUserID = userID
	return s.contrib, nil
}

func newLoyaltyTestHandler(repo *stubLoyaltyRepo) http.Handler {
	uc := usecase.NewLoyaltyQueryUseCase(repo)
	loyaltyHandler := NewLoyaltyHandler(uc)
	return NewRouter(
		NewHealthHandler(fakeHealthUseCase{}),
		NewAuthHandler(nil),
		NewProfileHandler(nil),
		NewRankingHandler(usecase.NewRankingUseCase(nil)),
		loyaltyHandler,
		nil,
		JWTMiddleware(stubTokenService{}),
	)
}

func TestLoyaltyHandler_OK(t *testing.T) {
	repo := &stubLoyaltyRepo{
		balance: 25,
		history: []entities.LoyaltyTransaction{
			{ID: "t1", Points: 10, Reason: "receipt_confirmed"},
			{ID: "t2", Points: 15, Reason: "receipt_confirmed;data_completion"},
		},
	}
	router := newLoyaltyTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/loyalty", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp usecase.LoyaltyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Balance != 25 {
		t.Errorf("Balance = %d, want 25", resp.Balance)
	}
	if len(resp.History) != 2 {
		t.Fatalf("expected 2 movements, got %d", len(resp.History))
	}
}

func TestLoyaltyHandler_Empty(t *testing.T) {
	repo := &stubLoyaltyRepo{}
	router := newLoyaltyTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/loyalty", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp usecase.LoyaltyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Balance != 0 {
		t.Errorf("Balance = %d, want 0", resp.Balance)
	}
	if resp.History == nil {
		t.Errorf("History must be a non-nil empty slice")
	}
	if len(resp.History) != 0 {
		t.Errorf("len(History) = %d, want 0", len(resp.History))
	}
}

func TestLoyaltyHandler_Unauthenticated(t *testing.T) {
	repo := &stubLoyaltyRepo{}
	router := newLoyaltyTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/loyalty", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if repo.lastUserID != "" {
		t.Errorf("repo must NOT be queried when unauthenticated; saw userID %q", repo.lastUserID)
	}
}

func TestLoyaltyHandler_CrossUserIsolation(t *testing.T) {
	repo := &stubLoyaltyRepo{balance: 99, history: []entities.LoyaltyTransaction{{ID: "t-x", Points: 99}}}
	router := newLoyaltyTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/loyalty", nil)
	// Do NOT send Authorization — use the legacy X-User-ID dev header to
	// act as user X. JWTMiddleware accepts this header for dev/tests.
	req.Header.Set("X-User-ID", "user-x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// The repo MUST have been queried with "user-x" (the identity present on the request),
	// never with any other user id hardcoded in the handler.
	if repo.lastUserID != "user-x" {
		t.Errorf("repository was queried with userID %q, want %q (cross-user isolation: only the request identity is used)", repo.lastUserID, "user-x")
	}
}

func TestLoyaltyHandler_MethodNotAllowed(t *testing.T) {
	repo := &stubLoyaltyRepo{}
	router := newLoyaltyTestHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/loyalty", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
