---
name: go-backend
description: Go backend development for AhorraApp — Clean Architecture (Constitution Article I), chi router, pgx/v5, JWT auth, port/adapter pattern, testing conventions, and project-specific helpers (writeJSON, writeError, userIDFromRequest, caarlos0/env config)
license: MIT
compatibility: opencode
metadata:
  audience: developers
  language: go
  framework: gin/chi
  project: ahorrapp
  architecture: clean-architecture
---

## What I do

- Implement Go code that respects AhorraApp's Constitution Article I (Clean Architecture with inward dependencies).
- Add new use cases, HTTP handlers, and Postgres adapters following the existing port/adapter pattern.
- Wire new endpoints into `cmd/api/main.go` and `internal/adapter/http/router.go` consistently.
- Write unit tests with fake repositories (table-driven) and HTTP integration tests with `httptest`.
- Add new migrations following the existing `migrations/00000N_name.up.sql` / `.down.sql` pattern.
- Surface new env-driven configuration through `internal/config/config.go` using `caarlos0/env/v11` struct tags.

## When to use me

Use this skill when:
- Creating or modifying any file under `internal/`, `cmd/`, or `migrations/` in the Go backend.
- Adding a new REST endpoint under `/api/v1/...`.
- Adding a new domain entity, port interface, use case, or repository.
- Adding a new migration (SQL pair).
- Adding a new environment variable to the config.
- Writing or fixing Go unit / integration tests in `internal/`.
- Touching `cmd/api/main.go` for dependency wiring.

## Architecture (Constitution Article I)

Inward dependency direction only:

```
HTTP handlers (adapter)  →  use cases (application)  →  entities (domain)
        ↑                          ↑
        └── ports (interfaces) ────┘   (defined in domain, implemented in adapters)
```

**The domain MUST NOT import:** `pgx`, `chi`, `minio`, `redis`, `ocr-service`, or any adapter package. If a domain file imports infrastructure, it is a constitution violation.

### Layered file layout

```
cmd/api/main.go                       # entry point, DI wiring only
internal/
  config/config.go                    # env config (caarlos0/env)
  domain/
    entities/                         # pure types: Receipt, User, PriceAggregate, ...
    ports/                            # interfaces: ReceiptRepository, LoyaltyRepository, ...
  usecase/                            # application logic: receipt_confirm.go, loyalty_award.go, ...
  adapter/
    http/                             # chi handlers, router, helpers (writeJSON, writeError, userIDFromRequest)
    postgres/                         # pgx implementations of ports
migrations/00000N_short_name.{up,down}.sql
```

## Ports pattern

Every external dependency is accessed through an interface in `internal/domain/ports/`. The interface lives with the domain; the implementation lives in the adapter layer.

```go
// internal/domain/ports/loyalty_repository.go
package ports

import "context"

type LoyaltyRepository interface {
    AwardForReceipt(ctx context.Context, userID, receiptID string, points int, reason string) error
    DailyGrantCount(ctx context.Context, userID string) (int, error)
    Balance(ctx context.Context, userID string) (int, error)
    History(ctx context.Context, userID string, limit int) ([]entities.LoyaltyTransaction, error)
}
```

The use case depends on the interface, never the concrete type:

```go
// internal/usecase/loyalty_award.go
type LoyaltyAwardUseCase struct {
    repo ports.LoyaltyRepository
    // ...
}
```

`cmd/api/main.go` is the only place that knows about both sides; it wires `postgres.NewLoyaltyRepository(pool)` into the use case.

**Two named ports are mandatory** (Constitution Art. I.4):
- `OCRProvider` — swap self-hosted PaddleOCR for a paid API by changing the adapter only.
- `StorageProvider` — same S3 adapter targets MinIO locally and Hetzner Object Storage in prod.

Any new external dependency (e.g., email, push notifications) must follow the same pattern.

## HTTP handler conventions

### File structure for a feature

```go
// internal/adapter/http/loyalty_handler.go
package httpapi

import (
    "net/http"
    "ahorrapp/internal/usecase"
)

type LoyaltyHandler struct {
    query *usecase.LoyaltyQueryUseCase
}

func NewLoyaltyHandler(query *usecase.LoyaltyQueryUseCase) *LoyaltyHandler {
    return &LoyaltyHandler{query: query}
}

func (h *LoyaltyHandler) GetLoyalty(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }
    userID := userIDFromRequest(r)
    if userID == "" {
        writeError(w, http.StatusUnauthorized, "invalid or expired token")
        return
    }

    resp, err := h.query.GetLoyalty(r.Context(), userID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, resp)
}
```

### Shared helpers (in `internal/adapter/http/auth_handler.go`)

- `writeJSON(w, status, body any)` — sets `Content-Type: application/json`, encodes body.
- `writeError(w, status, message)` — encodes `{"error": "<message>"}`.
- `userIDFromRequest(r)` — extracts user ID from `X-User-ID` header (dev/test) or context (prod via JWT middleware). Returns `""` if neither is set.
- `WithUserID(r, userID)` — attaches the user ID to the request context.

Always use these helpers. Never call `json.NewEncoder` or set headers directly in a handler.

### Router registration

Mount the new handler in `internal/adapter/http/router.go`:

- Public endpoints: directly under `v1` (e.g., `/health`, `/auth/register`, `/auth/login`).
- Authenticated endpoints: inside the `v1.Group(func(authed chi.Router) { authed.Use(jwtMiddleware) ... })` block (e.g., `/me/loyalty`, `/ranking/products/search`).
- Receipt-related routes: call `registerReceiptRoutes(authed)` from the `ReceiptHandler.RegisterRoutes` method to keep them grouped.

### Error handling

- `400` — invalid input (missing query param, malformed UUID, validation failure).
- `401` — missing or invalid JWT (return early with `writeError(w, http.StatusUnauthorized, "invalid or expired token")`).
- `404` — resource not found.
- `500` — wrap and return the error message; do not leak internal details to clients.
- For string-matching error mapping (legacy): `if strings.Contains(err.Error(), "not found")` → 404. Prefer typed errors when adding new code.

## Use case conventions

```go
// internal/usecase/loyalty_award.go
type LoyaltyAwardUseCase struct {
    repo                  ports.LoyaltyRepository
    basePoints            int
    firstObservationBonus int
    dataCompletionBonus   int
    dailyAwardCap         int
}

func NewLoyaltyAwardUseCase(
    repo ports.LoyaltyRepository,
    basePoints, firstObsBonus, dataCompletionBonus, dailyCap int,
) *LoyaltyAwardUseCase {
    return &LoyaltyAwardUseCase{...}
}
```

- Constructor takes **only** ports and primitive config values (not other use cases, not loggers from outside).
- A use case may depend on other use cases by interface (e.g., `ConfirmReceipt` may call `LoyaltyAwardUseCase.AwardForReceipt`), but the dependency is injected in `main.go`, not constructed inline.
- **Idempotency lives at the data layer** (unique constraints, partial unique indexes). The use case swallows the corresponding `ports.ErrAlreadyAwarded` (or analogous) error and treats it as a no-op.
- Errors that should not break a user-facing flow (e.g., loyalty award failure during confirm) are logged with `log.Printf` and swallowed. The receipt is still confirmed.

## Testing conventions

### Unit tests for use cases

Use **fake repositories** defined at the top of the test file (no external mocking library):

```go
// internal/usecase/loyalty_award_test.go
type fakeLoyaltyRepo struct {
    awardedCalls  []awardCall
    dailyCount    int
    awardErr      error
    // ...
}

func (f *fakeLoyaltyRepo) AwardForReceipt(_ context.Context, userID, receiptID string, points int, reason string) error {
    f.awardedCalls = append(f.awardedCalls, awardCall{userID, receiptID, points, reason})
    return f.awardErr
}
// ... implement every port method

func TestLoyaltyAward_BaseOnceAwarded(t *testing.T) {
    repo := &fakeLoyaltyRepo{}
    uc := awardUseCaseBase(repo)
    // ...
}
```

- Standard library `testing` only (no testify, no gomock).
- **Table-driven tests** for use cases with multiple scenarios.
- Assertions use plain `if`/`t.Errorf` (or `t.Fatalf` for setup), not testify.

### HTTP integration tests

Use `httptest.NewRecorder` and a hand-rolled router:

```go
// internal/adapter/http/loyalty_handler_test.go
func TestLoyaltyHandler_GetLoyalty_Unauthorized(t *testing.T) {
    h := NewLoyaltyHandler(...)
    req := httptest.NewRequest(http.MethodGet, "/api/v1/me/loyalty", nil)
    rr := httptest.NewRecorder()
    h.GetLoyalty(rr, req)
    if rr.Code != http.StatusUnauthorized { ... }
}
```

For full routing tests, use the helpers in `internal/adapter/http/router_test_helpers_test.go` (builds a router with stubbed dependencies).

## Configuration pattern

Add new env vars to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields
    LoyaltyBasePoints            int `env:"LOYALTY_BASE_POINTS" envDefault:"10"`
    LoyaltyFirstObservationBonus int `env:"LOYALTY_FIRST_OBSERVATION_BONUS" envDefault:"5"`
    // ...
}

func Load() (Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return Config{}, err
    }
    // cross-field validation lives here, e.g. ServerPort range
    return cfg, nil
}
```

- Defaults are encoded in struct tags; never hardcoded in the use case.
- `envDefault:""` for optional strings; `envDefault:"0"` for optional numeric flags.
- Use `env:",required"` only when a missing value must abort startup (e.g., `DATABASE_URL`).
- Add a corresponding test in `internal/config/config_test.go` for any new validation rule.

Update `.env.example` to mirror any new variables.

## Migrations

```text
migrations/00000N_short_snake_name.up.sql
migrations/00000N_short_snake_name.down.sql
```

- Numbered sequentially (next is `000006`).
- Use `golang-migrate` SQL format: one `+migrate Up` / `+migrate Down` per file is unnecessary because files are split by direction; just put plain SQL in each.
- Always provide a down migration that reverses the up.
- New tables: include a primary key, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, and any foreign keys with `ON DELETE` rules declared explicitly.
- Partial unique indexes (e.g., `WHERE receipt_id IS NOT NULL`) are the idiomatic way to enforce idempotency for nullable columns.

## Wiring in `cmd/api/main.go`

When adding a new use case + handler + repository:

1. Construct the Postgres repository: `repo := postgres.NewXxxRepository(pool)`.
2. Construct the use case: `uc := usecase.NewXxxUseCase(repo, cfg.Xxx)`.
3. Construct the handler: `h := httpapi.NewXxxHandler(uc)`.
4. Pass the handler to `httpapi.NewRouter(...)`.
5. Register the route in `internal/adapter/http/router.go` (either as a top-level handler or as a sub-router).

Never instantiate a use case inside a handler. Never instantiate a repository inside a use case.

## Code style

- `gofmt` and `goimports` clean.
- No comments unless the meaning is non-obvious; comments are in English (Constitution Art. IX).
- Error messages start lowercase; no trailing punctuation.
- Use `%w` for error wrapping; surface wrapped errors with `errors.Is`/`errors.As` in tests.
- Prefer small, focused functions over long ones; use cases are the only place where business flow gets long.
- `log.Printf` for diagnostic logging; do not introduce a logger interface in domain code.

## API versioning (Constitution Art. IV)

- All new endpoints live under `/api/v1/...`.
- Breaking changes require a new path version (`/api/v2/...`), not a query parameter.
- Document every endpoint's request, response, and error shapes in `specs/NNN-feature/contracts/`.
- The Flutter app consumes ONLY published contracts; new request/response fields must be added to the contract first.
