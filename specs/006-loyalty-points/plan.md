# Implementation Plan: Loyalty Points for Receipt Upload

**Branch**: `006-loyalty-points` | **Date**: 2026-06-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-loyalty-points/spec.md`

## Summary

On receipt confirmation, the user earns a configurable base amount of loyalty
points plus optional "first observation" and "data completion" bonuses. Every
movement is recorded in `loyalty_transactions`. A new endpoint
`GET /api/v1/me/loyalty` returns the authenticated user's point balance and
the full ordered history. Anti-abuse is enforced twice: a unique
`(receipt_id)` constraint on `loyalty_transactions` makes double-award
impossible at the data level, and a configurable daily cap on
point-granting receipts causes confirmations beyond the cap to succeed (and
feed the price engine) but award zero points with a recorded reason. The
existing `users.points` column remains as a denormalized cache that equals
`SUM(loyalty_transactions.points)`.

## Technical Context

**Language/Version**: Go 1.22+ (existing backend).

**Primary Dependencies**: `go-chi/chi/v5` (HTTP router), `jackc/pgx/v5`
(Postgres driver), `caarlos0/env/v11` (config), `golang-jwt/jwt` (auth).

**Storage**: PostgreSQL 16. The `loyalty_transactions` table already exists
(migration `000003`) with columns `id, user_id, points, reason, created_at`.
This feature adds an optional `receipt_id` column with a UNIQUE constraint
to enforce idempotency and a `daily_grant_count` lookup index.

**Testing**: `testing` (stdlib). Existing pattern: table-driven unit tests
in `internal/usecase/*_test.go` using fake repositories
(`internal/usecase/receipt_confirm_test.go`); HTTP integration tests in
`internal/adapter/http/*_handler_test.go` using `httptest`.

**Target Platform**: Linux server (Docker, local-first per Constitution Art. VI).

**Project Type**: web-service (REST/JSON backend), part of a mobile + API app.

**Performance Goals**: loyalty endpoint returns in under 1 second for a
typical user (SC-006).

**Constraints**: Clean Architecture (Constitution Art. I) — domain must not
import infra. JWT auth (Art. VII).

**Scale/Scope**: MVP scale, single VPS, thousands of users, tens of
transactions per user.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

This plan satisfies every applicable constitution article. Cited below with
the specific design decision that satisfies each.

- **Article I — Clean Architecture (inward dependencies, replaceable ports).**
  The new point-awarding logic lives in a use case
  (`internal/usecase/loyalty_award.go`) that depends only on a new
  `ports.LoyaltyRepository` interface (defined in `internal/domain/ports/`)
  and the existing `ports.ReceiptRepository` / `ports.UserRepository`. No
  infra import reaches the domain. The HTTP layer adds a `LoyaltyHandler`
  in `internal/adapter/http/` consuming the use case; the Postgres adapter
  implements the new port in `internal/adapter/postgres/`. ✓
- **Article II — Spec first, code second.** Spec approved at
  `specs/006-loyalty-points/spec.md`; this plan is the HOW. ✓
- **Article III — Tests.** Unit tests for the new use case using fake
  repositories (TDD). Integration test for the new endpoint verifying
  auth rejection and balance/history correctness. ✓
- **Article IV — Explicit versioned contracts.** The new endpoint is
  `/api/v1/me/loyalty` (versioned, v1). Its request/response/error shapes
  are documented in `contracts/me-loyalty.md` (Phase 1). ✓
- **Article V — Data, currency, normalization.** Loyalty points are
  currency-agnostic; the user earns points per *confirmed receipt*, never
  per currency. First-observation detection uses the canonical
  `PriceObservation(product_id, store_id)` pairs that already exist after
  normalization. No currency mixing concern applies to points. ✓
- **Article VI — Simplicity, cost, local-first.** No new managed service,
  no new container, no paid dependency. Configuration is via environment
  variables (`LOYALTY_*`) following the existing `config.go` pattern. ✓
- **Article VII — Security.** The endpoint runs behind the existing
  `JWTMiddleware` already mounted on `/api/v1` authenticated routes;
  unauthenticated access is rejected. Input is the user's own identity from
  the JWT — no user-controlled query parameter that could leak other users'
  data. ✓
- **Article VIII — Ready to grow, no over-engineering.** Redeeming points,
  leaderboards, and reward tiers are explicitly out of scope (spec
  Assumptions). The `loyalty_transactions` schema and the
  `reason` field allow future redemption movements (negative points) without
  schema change. ✓
- **Article IX — English.** All new code, comments (only where necessary),
  contracts, and docs are in English. ✓

No unconstitutional deviations; Complexity Tracking table is empty.

## Project Structure

### Documentation (this feature)

```text
specs/006-loyalty-points/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── me-loyalty.md
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

This feature extends the existing Go Clean Architecture layout. No new
top-level directories. Only new files inside the existing structure and one
new migration pair.

```text
internal/
├── config/
│   └── config.go                      # add LOYALTY_* fields (edit)
├── domain/
│   ├── entities/
│   │   ├── loyalty_transaction.go      # new: extend with ReceiptID, helpers
│   │   └── user.go                    # edit: drop nothing, keep Points
│   └── ports/
│       ├── loyalty_repository.go       # new interface: AwardTransactions, History, Balance, DailyGrantCount
│       ├── first_observation.go        # new interface: check (product,store) and store first observation
│       └── user_repository.go          # edit: keep GetPoints + RecentTransactions for backward compat
├── usecase/
│   ├── loyalty_award.go                # new: AwardForReceipt(ctx, receipt, observations, user) ([]LoyaltyTransaction, error)
│   ├── loyalty_award_test.go           # new: unit tests
│   ├── loyalty_query.go                # new: GetLoyalty(ctx, userID) (Balance, History)
│   ├── loyalty_query_test.go           # new
│   └── receipt_confirm.go              # edit: invoke LoyaltyAwardUseCase after recompute (idempotent)
├── adapter/
│   ├── http/
│   │   ├── loyalty_handler.go          # new GET /me/loyalty
│   │   ├── loyalty_handler_test.go     # new
│   │   └── router.go                   # edit: register new route under authed group
│   └── postgres/
│       ├── loyalty_repository.go       # new impl
│       ├── first_observation_repository.go # new impl, uses price_observations/store
│       └── receipt_repository.go       # edit: ConfirmReceipt now also returns storeID + new observation product_ids for the award use case (or alternative accessible via the persistence call). Decision: keep ConfirmReceipt unchanged; expand its return signature to surface the persisted observations + storeID for the award step to detect first observations.
└── ...

migrations/
├── 000005_loyalty_receipt_link.up.sql   # new
└── 000005_loyalty_receipt_link.down.sql # new

cmd/api/main.go                          # edit: wire new ports + use cases + handler
```

**Structure Decision**: Extend the existing single-binary Go API in
`internal/` (the only project module). No multi-module split. No Flutter
changes in this feature — the mobile app already consumes
`/api/v1/users/me/points`; an alias or migration to `/api/v1/me/loyalty` is
documented in the contract (the two endpoints coexist during the transition;
only `/me/loyalty` is the canonical one going forward, with
`/users/me/points` left as a backward-compatible alias returning the same
shape).

## Complexity Tracking

> Empty — no constitution violations to justify.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |