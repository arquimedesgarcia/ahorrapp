package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahorrapp/internal/usecase"
)

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
	router := NewRouter(h)
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
