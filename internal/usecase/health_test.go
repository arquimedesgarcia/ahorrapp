package usecase

import (
	"context"
	"errors"
	"testing"
)

type mockChecker struct {
	name string
	err  error
}

func (m mockChecker) Name() string { return m.name }
func (m mockChecker) Check(context.Context) error {
	return m.err
}

func TestHealthUseCase_AllHealthy(t *testing.T) {
	uc := NewHealthUseCase(
		mockChecker{name: "postgres"},
		mockChecker{name: "redis"},
	)

	resp := uc.Execute(context.Background())
	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %s", resp.Status)
	}
	if !resp.Dependencies["postgres"].Reachable || !resp.Dependencies["redis"].Reachable {
		t.Fatalf("expected all dependencies reachable: %#v", resp.Dependencies)
	}
}

func TestHealthUseCase_Degraded(t *testing.T) {
	uc := NewHealthUseCase(
		mockChecker{name: "postgres", err: errors.New("db down")},
		mockChecker{name: "redis"},
	)

	resp := uc.Execute(context.Background())
	if resp.Status != "degraded" {
		t.Fatalf("expected status degraded, got %s", resp.Status)
	}
	if resp.Dependencies["postgres"].Reachable {
		t.Fatalf("expected postgres unreachable")
	}
	if resp.Dependencies["postgres"].Error == "" {
		t.Fatalf("expected postgres error message")
	}
}
