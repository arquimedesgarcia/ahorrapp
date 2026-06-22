package usecase

import (
	"context"
	"time"

	"ahorrapp/internal/domain/ports"
)

type DependencyStatus struct {
	Name      string `json:"name"`
	Reachable bool   `json:"reachable"`
	Error     string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status       string                      `json:"status"`
	Dependencies map[string]DependencyStatus `json:"dependencies"`
}

type HealthUseCase struct {
	checkers []ports.DependencyChecker
}

func NewHealthUseCase(checkers ...ports.DependencyChecker) *HealthUseCase {
	return &HealthUseCase{checkers: checkers}
}

func (u *HealthUseCase) Execute(ctx context.Context) HealthResponse {
	resp := HealthResponse{
		Status:       "ok",
		Dependencies: make(map[string]DependencyStatus, len(u.checkers)),
	}

	for _, checker := range u.checkers {
		depCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := checker.Check(depCtx)
		cancel()

		status := DependencyStatus{Name: checker.Name(), Reachable: err == nil}
		if err != nil {
			status.Error = err.Error()
			resp.Status = "degraded"
		}
		resp.Dependencies[checker.Name()] = status
	}

	return resp
}
