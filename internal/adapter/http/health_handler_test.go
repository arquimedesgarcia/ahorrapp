package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ahorrapp/internal/domain/ports"
	"ahorrapp/internal/usecase"
)

type stubTokenService struct{}

func (stubTokenService) Generate(string, string) (string, error) { return "stub", nil }
func (stubTokenService) Validate(string) (*ports.TokenClaims, error) {
	return &ports.TokenClaims{UserID: "test-user", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type fakeHealthUseCase struct {
	response usecase.HealthResponse
}

func (f fakeHealthUseCase) Execute(context.Context) usecase.HealthResponse {
	return f.response
}

func TestHealthHandler_ReturnsStatusAndDependencies(t *testing.T) {
	uc := fakeHealthUseCase{response: usecase.HealthResponse{
		Status: "ok",
		Dependencies: map[string]usecase.DependencyStatus{
			"postgres": {Name: "postgres", Reachable: true},
			"redis":    {Name: "redis", Reachable: true},
		},
	}}

	h := NewHealthHandler(uc)
	router := NewRouter(
		h,
		NewAuthHandler(nil),
		NewProfileHandler(nil),
		NewRankingHandler(),
		nil,
		JWTMiddleware(stubTokenService{}),
	)
	ts := httptest.NewServer(router)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("get health: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var got usecase.HealthResponse
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != "ok" {
		t.Fatalf("expected status ok, got %s", got.Status)
	}
	if !got.Dependencies["postgres"].Reachable {
		t.Fatalf("expected postgres reachable")
	}
}
