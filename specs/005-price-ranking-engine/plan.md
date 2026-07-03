# Implementation Plan: Average-Price Engine and "Where to Buy Cheaper" Ranking

**Branch**: `005-price-ranking-engine` | **Date**: 2025-06-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-price-ranking-engine/spec.md`

## Summary

Build the average-price aggregation engine and ranking endpoints that turn
confirmed receipt observations into actionable "where to buy cheaper"
recommendations. When a receipt is confirmed, the system synchronously
recomputes a precomputed `price_aggregates` cache per
(product, store, currency) triple, applying a configurable age-threshold
filter to exclude stale observations. Two authenticated endpoints serve
the ranking: `GET /api/v1/products/{id}/prices` (per-product ranking)
and `GET /api/v1/search?q=...` (product search with cheapest store per
currency). The existing stub `RankingHandler` is replaced with a real
implementation wired to a new `RankingUseCase` and
`PriceAggregateRepository`. An optional PostGIS proximity filter is
included as a thin, opt-in query path.

## Technical Context

**Language/Version**: Go 1.23 (existing backend)

**Primary Dependencies**: go-chi/chi v5 (HTTP router), jackc/pgx v5
(Postgres driver), redis/go-redis v9, golang-migrate (migrations),
caarlos0/env (config). No new external dependencies needed.

**Storage**: PostgreSQL 16 with PostGIS (already active in Docker).
New table `price_aggregates`. New columns `lat`/`long` on `stores`.
New `unaccent` extension for accent-insensitive search. Redis remains
for the OCR queue only — no ranking cache in Redis at MVP (YAGNI,
Article VI.1).

**Testing**: Go `testing` package with table-driven unit tests for use
cases; `net/http/httptest` for handler integration tests. Existing test
helpers in `internal/adapter/http/router_test_helpers_test.go` are
reused.

**Target Platform**: Linux container (Docker), same as existing API.

**Project Type**: Go web service (Clean Architecture), existing project.

**Performance Goals**: Recompute aggregates for a 20-line-item receipt
within the confirmation transaction (< 200 ms at MVP scale). Ranking
queries return in < 500 ms for MVP-scale observation counts (< 10k
rows). No hard SLO beyond "imperceptible to the user" (SC-006).

**Constraints**: Currency isolation is mandatory (constitution Article
V.1); no average may mix currencies. Age threshold is configurable via
`PRICE_AGE_THRESHOLD_DAYS` env var (FR-004). All new endpoints require
JWT auth (FR-012). Versioned contract under `/api/v1/` (Article IV.1).
The Flutter app's existing ranking contract must not break (Article
IV.4).

**Scale/Scope**: MVP single-tenant, tens of receipts, thousands of
observations. No sharding, no read replicas, no Redis cache for
aggregates. Proximity filtering is optional and opt-in (P3).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Article | Requirement | How this plan satisfies it |
|---------|-------------|----------------------------|
| **I.1** | Layered dependencies point inward (handlers → usecases → entities). | New `RankingUseCase` lives in `internal/usecase`, depends only on `ports.RankingRepository` (interface in `internal/domain/ports`). HTTP handler in `internal/adapter/http` calls the use case. No layer violations. |
| **I.2** | Domain MUST NOT import frameworks. | `PriceAggregate` entity and `RankingRepository` interface are pure Go in `internal/domain`. The Postgres adapter imports `pgx`, not the domain. |
| **I.3** | External deps via ports. | `RankingRepository` interface in domain; `postgres.PriceAggregateRepository` adapter implements it. |
| **I.4** | Replaceable details. | No new external dependency introduced. PostGIS is accessed via the existing Postgres adapter, behind the repository port. |
| **II.1** | No code without approved spec + plan. | This plan is the approved plan; spec is `specs/005-price-ranking-engine/spec.md`. |
| **II.3** | Plan cites constitution articles. | This table is the citation. |
| **III.1** | Every domain use case MUST have unit tests. | `RankingUseCase` and `PriceAggregateRecomputeUseCase` will have unit tests with stub repositories. |
| **III.2** | Critical endpoints (price ranking) MUST have integration tests. | `GET /api/v1/products/{id}/prices` and `GET /api/v1/search` will have `httptest` integration tests. |
| **IV.1** | Versioned contracts under `/api/v1/`. | New endpoints under `/api/v1/products/{id}/prices` and `/api/v1/search`. |
| **IV.2** | Every endpoint documents request/response/error. | Contracts written in `specs/005-price-ranking-engine/contracts/`. |
| **IV.4** | Flutter consumes ONLY published contracts. | Existing `ranking-api-contract.md` honored; no breaking change to the Flutter app. |
| **V.1** | Currency never mixed. | `price_aggregates` key includes `currency`; every query groups by currency. Research R5. |
| **V.2** | Store carries geolocation. | Migration adds `lat`/`long` to `stores`; proximity query uses PostGIS. Research R6. |
| **V.3** | Product names normalized before averaging. | Existing `normalizeProductTx` reused; search uses `ILIKE` + `unaccent()`. Research R4. |
| **VI.1** | Simplest solution (YAGNI). | Hard-filter age threshold (not weighted); synchronous recompute (not async); no Redis cache for aggregates. Research R1, R2, R3. |
| **VI.4** | Local-first. | All changes run in the existing Docker stack; no cloud dependency. |
| **VII.1** | JWT auth. | Both new endpoints behind existing `JWTMiddleware`. |
| **VII.2** | Input validation. | Search `q` min length validated; product ID format validated; age threshold bounded. |
| **IX.1** | English for all artifacts. | All specs, contracts, code, comments in English. |

**Gate result**: PASS. No violations. No complexity tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/005-price-ranking-engine/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── product-prices-api-contract.md
│   └── product-search-api-contract.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
migrations/
├── 000004_price_aggregates_store_geo.up.sql      # New table + store geo columns
└── 000004_price_aggregates_store_geo.down.sql

internal/domain/
├── entities/
│   └── price_aggregate.go          # New entity: PriceAggregate
└── ports/
    └── ranking.go                  # New port: RankingRepository

internal/usecase/
├── price_aggregate_recompute.go    # Recompute aggregates on confirm
├── price_aggregate_recompute_test.go
├── ranking.go                      # RankingUseCase (per-product + search)
└── ranking_test.go

internal/adapter/
├── postgres/
│   ├── price_aggregate_repository.go  # Implements ports.RankingRepository
│   └── price_aggregate_repository_test.go
└── http/
    ├── ranking_handler.go              # Replaces stub with real impl
    └── ranking_handler_test.go

internal/config/
└── config.go            # Add PriceAgeThresholdDays field

cmd/api/main.go          # Wire new repository + use case + handler
```

**Structure Decision**: The plan follows the existing Clean Architecture
layout. New entity in `internal/domain/entities`, new port in
`internal/domain/ports`, new use case in `internal/usecase`, new
Postgres adapter in `internal/adapter/postgres`, and the existing stub
`RankingHandler` in `internal/adapter/http` is replaced with the real
implementation. No new packages or directories are created beyond what
the architecture already defines.

## Complexity Tracking

> Not applicable — no constitution violations to justify.