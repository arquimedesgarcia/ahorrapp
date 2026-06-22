# Implementation Plan: Backend Skeleton

**Branch**: `001-backend-skeleton` | **Date**: 2026-06-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-backend-skeleton/spec.md`

## Summary

Establish the Go backend foundation: Clean Architecture project layout, Docker Compose
orchestration of API + PostgreSQL 16/PostGIS + Redis + MinIO, a `GET /api/v1/health` endpoint
that reports per-dependency reachability, versioned database migrations via golang-migrate,
environment-variable-driven configuration, and multi-stage Docker build producing a minimal
binary. The domain layer defines ports (interfaces) for every external dependency; no concrete
infrastructure imports reach the domain. Exactly the skeleton — zero auth, zero business logic.

## Technical Context

**Language/Version**: Go 1.23+ (latest stable as of 2026)
**Primary Dependencies**: chi (HTTP router), pgx (PostgreSQL driver), go-redis (Redis client),
  golang-migrate (migrations), minio-go (S3 client, placeholder only), caarlos0/env (env parsing)
**Storage**: PostgreSQL 16 + PostGIS (via pgx); MinIO (S3-compatible, local dev)
**Testing**: Go standard library `testing` + `httptest`; no external test framework needed for
  this skeleton's scope
**Target Platform**: Linux containers (Docker); dev on Windows/macOS/Linux via Docker Desktop
**Project Type**: Web service (REST/JSON API)
**Performance Goals**: Health endpoint <2 s (including DB + Redis round-trips); cold start
  <60 s with `docker compose up`
**Constraints**: Zero cloud accounts, zero paid services during development; domain MUST NOT
  reference any infrastructure package; all secrets via env vars
**Scale/Scope**: Single developer (MVP foundation); single HTTP endpoint; one migration table

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Article | Requirement | How this plan satisfies it | Status |
|---------|-------------|----------------------------|--------|
| **Art. I** — Clean Architecture | Domain must not import frameworks/DB/HTTP; external deps behind ports | `/internal/domain/ports/` defines interfaces (`HealthChecker`, `Repository`, `StorageProvider`, `OCRProvider`, `Cache`); adapters in `/internal/adapter/` implement them. Domain import graph: `domain/ports` ← `usecase` ← `adapter/*` ← `cmd`. No `net/http`, `pgx`, `go-redis`, or `minio-go` in `domain/` or `usecase/`. | PASS |
| **Art. I.4** — Named replaceable ports | `OCRProvider` and `StorageProvider` must exist and be the only way to reach their concern | Ports defined as Go interfaces in `domain/ports/` with NO concrete implementation in the skeleton. Adapters (`internal/adapter/ocr/`, `internal/adapter/storage/`) are empty stubs returning "not implemented" — the interfaces exist, the swap contract is ready, and nothing else in the system can reach OCR or storage without going through the port. | PASS |
| **Art. IV** — Explicit, versioned contracts | `/api/v1/` versioning; OpenAPI preferred; Flutter consumes only published contracts | Health endpoint at `GET /api/v1/health`. Response schema documented in `contracts/health-contract.md`. OpenAPI spec deferred to later epics when endpoint count grows beyond 3-4. | PASS |
| **Art. VI** — Simplicity, cost, local-first | Simplest solution, Docker-portable, zero paid services, local-first | Single binary, chi (lightweight router), `pgx` (no ORM), `docker compose up` runs everything locally. No paid managed services. `.env` file with sensible defaults for local dev. | PASS |
| **Art. III** — Tests | Unit tests for use cases; integration test for critical endpoint | `usecase/health_test.go` unit-tests the health-check logic with mock ports. `adapter/http/health_handler_test.go` integration-tests the `/health` endpoint through `httptest`. | PASS |
| **Art. II** — Spec first | No code without approved spec + plan | ✅ This plan cites the approved [spec.md](./spec.md). | PASS |
| **Art. V** — Data, currency | Not applicable to skeleton (no PriceObservation entities yet) | N/A — deferred to receipt/price-engine epics. PostGIS is enabled in this skeleton so the column type exists when needed. | N/A |
| **Art. VII** — Minimal security | JWT, bcrypt, rate limiting | Not applicable to skeleton — auth is E2. The skeleton includes no auth middleware. `FR-005` (env vars for secrets) is the only security-relevant requirement at this stage. | N/A |
| **Art. VIII** — Ready to grow | Architecture allows future additions without core rewrite | Ports for `OCRProvider`, `StorageProvider`, and `Repository` are declared as interfaces now; adapters are stubs. Adding real implementations later requires no changes to domain or use cases. | PASS |
| **Art. IX** — Working language | English for code, identifiers, comments, docs | All identifiers and documentation are in English. | PASS |

**Gate result**: PASS — zero violations, zero exceptions to justify. Constitution Check re-validated after Phase 1 design: still PASS.

## Project Structure

### Documentation (this feature)

```text
specs/001-backend-skeleton/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: technology research & decisions
├── data-model.md        # Phase 1: entities, configuration schema
├── quickstart.md        # Phase 1: validation runbook
├── contracts/           # Phase 1: API contracts
│   └── health-contract.md
└── tasks.md             # Phase 2: /speckit.tasks output (NOT created here)
```

### Source Code (repository root)

```text
ahorrapp/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point: wire config, adapters, use cases, start server
├── internal/
│   ├── config/
│   │   └── config.go                # Env-var loading into typed Config struct
│   ├── domain/
│   │   ├── ports/
│   │   │   ├── health.go            # HealthChecker interface (ping DB, ping Redis)
│   │   │   ├── repository.go        # Repository port (placeholder)
│   │   │   ├── storage.go           # StorageProvider port (S3, placeholder)
│   │   │   ├── ocr.go               # OCRProvider port (placeholder)
│   │   │   └── cache.go             # Cache port (placeholder)
│   │   └── entities/                # Empty for this skeleton (entities come in later epics)
│   ├── usecase/
│   │   └── health.go                # HealthCheck use case: orchestrates per-port checks
│   └── adapter/
│       ├── http/
│       │   ├── router.go            # chi router setup, middleware, route registration
│       │   ├── health_handler.go    # GET /api/v1/health handler
│       │   └── health_handler_test.go
│       ├── postgres/
│       │   ├── client.go            # pgx pool creation from config
│       │   └── health_check.go      # Adapter: HealthChecker.PostgresCheck()
│       ├── redis/
│       │   ├── client.go            # go-redis client creation from config
│       │   └── health_check.go      # Adapter: HealthChecker.RedisCheck()
│       ├── storage/
│       │   └── client.go            # MinIO client stub (implements StorageProvider port)
│       └── ocr/
│           └── client.go            # OCR client stub (implements OCRProvider port)
├── internal/
│   └── usecase/
│       └── health_test.go           # Unit tests for HealthCheck use case with mock adapters
├── migrations/
│   ├── 000001_create_health_table.up.sql
│   └── 000001_create_health_table.down.sql
├── Dockerfile                       # Multi-stage: build (Go) → run (scratch/distroless)
├── docker-compose.yml               # Services: api, postgres, redis, minio
├── .env.example                     # Template with local dev defaults (no secrets)
├── scripts/
│   └── migrate.sh                   # Helper: runs golang-migrate against DATABASE_URL
├── go.mod
├── go.sum
└── .gitignore
```

**Structure Decision**: Single Go module at repository root with Clean Architecture
internal layout (`cmd/` entry point, `internal/` for everything else). The `internal/`
tree enforces Go's own visibility boundary (other modules cannot import it). Domain
ports in `internal/domain/ports/`, use cases in `internal/usecase/`, adapters in
`internal/adapter/`. No `/pkg/` directory — the skeleton has no public library surface.

### File Creation Order

Dependency-ordered: config must load before adapters; ports must exist before use cases;
adapters must exist before wiring in `cmd/api/main.go`.

1. `go.mod` — module initialization
2. `.env.example` — environment contract (no deps)
3. `internal/config/config.go` — env loading (no internal deps, only stdlib + env lib)
4. `internal/domain/ports/*.go` — all five port interfaces (zero deps, pure Go interfaces)
5. `internal/adapter/postgres/client.go` — pgx pool (depends on config)
6. `internal/adapter/redis/client.go` — go-redis client (depends on config)
7. `internal/adapter/postgres/health_check.go` — implements port (depends on client + port)
8. `internal/adapter/redis/health_check.go` — implements port (depends on client + port)
9. `internal/adapter/storage/client.go` — MinIO stub (depends on port)
10. `internal/adapter/ocr/client.go` — OCR stub (depends on port)
11. `internal/usecase/health.go` — HealthCheck logic (depends on ports only)
12. `internal/usecase/health_test.go` — unit tests for use case (depends on use case + ports)
13. `internal/adapter/http/router.go` — chi setup (depends on config)
14. `internal/adapter/http/health_handler.go` — handler (depends on use case)
15. `internal/adapter/http/health_handler_test.go` — integration test (depends on handler)
16. `cmd/api/main.go` — wire everything (depends on all above)
17. `migrations/000001_create_health_table.up.sql` + `.down.sql`
18. `scripts/migrate.sh` — migration runner
19. `Dockerfile` — multi-stage build
20. `docker-compose.yml` — service orchestration
21. `.gitignore` — exclude `.env`, binaries, IDE files

## Complexity Tracking

> No violations — nothing to justify.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none)    | —          | —                                    |
