# Tasks: Average-Price Engine and "Where to Buy Cheaper" Ranking

**Input**: Design documents from `/specs/005-price-ranking-engine/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included per Constitution Article III (unit tests for use cases, integration tests for critical endpoints).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Go backend with Clean Architecture. Paths relative to repository root:
- Domain entities: `internal/domain/entities/`
- Domain ports: `internal/domain/ports/`
- Use cases: `internal/usecase/`
- Postgres adapters: `internal/adapter/postgres/`
- HTTP handlers: `internal/adapter/http/`
- Config: `internal/config/`
- Migrations: `migrations/`
- Entry point: `cmd/api/main.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Migration files and configuration for the price-ranking engine

- [X] T001 Create migration `migrations/000004_price_aggregates_store_geo.up.sql` with `price_aggregates` table, `stores` geo columns, PostGIS geography column, GiST index, and `unaccent` extension per data-model.md
- [X] T002 [P] Create migration `migrations/000004_price_aggregates_store_geo.down.sql` with reverse DDL (drop table, drop index, drop columns, drop extension)
- [X] T003 [P] Add `PriceAgeThresholdDays` field to `internal/config/config.go` with env var `PRICE_AGE_THRESHOLD_DAYS` and default 90

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain entity, port interface, and Postgres repository that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 [P] Create `PriceAggregate` entity in `internal/domain/entities/price_aggregate.go` with fields: ProductID, StoreID, Currency, AveragePrice, MinPrice, SampleCount, LastObservedAt, UpdatedAt and JSON tags per data-model.md
- [X] T005 [P] Create `StoreRankingEntry` and `ProductSearchResult` value objects in `internal/domain/entities/price_aggregate.go` for ranking response shapes (store name, branch, average_price, min_price, currency, sample_count, last_observed_at, distance_km)
- [X] T006 Create `RankingRepository` interface in `internal/domain/ports/ranking.go` with methods: `RecomputeAggregate(ctx, productID, storeID, currency string, ageThresholdDays int) error`, `GetProductRanking(ctx, productID string, opts RankingQueryOptions) ([]StoreRankingEntry, error)`, `SearchProducts(ctx, query string) ([]ProductSearchResult, error)`, `GetProductName(ctx, productID string) (string, error)`
- [X] T007 Create `RankingQueryOptions` struct in `internal/domain/ports/ranking.go` with optional fields: Lat, Long, RadiusKm for proximity query support
- [X] T008 [P] Write unit test for `PriceAggregate` entity JSON marshaling in `internal/domain/entities/price_aggregate_test.go`
- [X] T009 Implement `PriceAggregateRepository` in `internal/adapter/postgres/price_aggregate_repository.go` with `RecomputeAggregate` method: SELECT AVG/MIN/COUNT/MAX from `price_observations` filtered by age threshold, UPSERT into `price_aggregates` via `ON CONFLICT (product_id, store_id, currency) DO UPDATE` (depends on T001, T004, T006)
- [X] T010 [P] Implement `GetProductName` method in `internal/adapter/postgres/price_aggregate_repository.go`: SELECT canonical_name FROM products WHERE id = $1 (depends on T009)
- [X] T011 Write unit test for `PriceAggregateRepository.RecomputeAggregate` in `internal/adapter/postgres/price_aggregate_repository_test.go` using a test Postgres connection or mock pool (depends on T009)

**Checkpoint**: Foundation ready — entity, port, repository, and migration exist. User story implementation can now begin.

---

## Phase 3: User Story 1 — Recompute Averages After Receipt Confirmation (Priority: P1) 🎯 MVP

**Goal**: When a receipt is confirmed, the system synchronously recomputes the `PriceAggregate` for each affected (product, store, currency) triple within the same transaction.

**Independent Test**: Upload and confirm two receipts for the same product at the same store in the same currency with different unit prices. Query `price_aggregates` for that triple. Verify that the average, minimum, and sample count reflect both observations.

### Tests for User Story 1

- [X] T012 [P] [US1] Write unit test for `PriceAggregateRecomputeUseCase` in `internal/usecase/price_aggregate_recompute_test.go` with a stub `RankingRepository` — verify it calls `RecomputeAggregate` for each unique (productID, storeID, currency) triple in the observations list
- [X] T013 [P] [US1] Write unit test in `internal/usecase/price_aggregate_recompute_test.go` for currency isolation — verify two observations with different currencies produce two separate `RecomputeAggregate` calls
- [X] T014 [P] [US1] Write unit test in `internal/usecase/price_aggregate_recompute_test.go` for new product — verify `RecomputeAggregate` is called when no prior aggregate exists

### Implementation for User Story 1

- [X] T015 [US1] Create `PriceAggregateRecomputeUseCase` in `internal/usecase/price_aggregate_recompute.go` with method `Execute(ctx context.Context, observations []entities.PriceObservation, ageThresholdDays int) error` that deduplicates observations by (productID, storeID, currency) and calls `repo.RecomputeAggregate` for each triple (depends on T009, T012)
- [X] T016 [US1] Add `PriceAggregateRecompute` dependency to `ReceiptConfirmUseCase` in `internal/usecase/receipt_confirm.go` — after `repo.ConfirmReceipt` succeeds, call `recompute.Execute(ctx, observations, cfg.AgeThresholdDays)` (depends on T015)
- [X] T017 [US1] Update `cmd/api/main.go` to wire `PriceAggregateRepository` and `PriceAggregateRecomputeUseCase` into `ReceiptConfirmUseCase` constructor (depends on T015, T016)
- [X] T018 [US1] Update `internal/usecase/receipt_confirm_test.go` stub to include the new `PriceAggregateRecompute` dependency in the `confirmRepoStub` (depends on T016)
- [X] T019 [US1] Rebuild and restart API container: `docker compose up -d --build api` then verify migration 000004 ran with `docker exec ahorrapp-postgres psql -U ahorrapp -d ahorrapp -c "\d price_aggregates"` (depends on T001, T017)

**Checkpoint**: User Story 1 is functional — confirming a receipt populates `price_aggregates`. Test independently via quickstart Scenario 1.

---

## Phase 4: User Story 2 — View Per-Store Ranking for a Specific Product (Priority: P1)

**Goal**: Expose `GET /api/v1/products/{id}/prices` that returns stores ranked cheapest-first per currency for a given product.

**Independent Test**: Seed `price_aggregates` directly via SQL for product P at stores S1 ($1.00), S2 ($1.50), S3 ($1.20). Query the endpoint. Verify stores are ordered S1, S3, S2 and each entry includes average_price, min_price, sample_count, last_observed_at.

### Tests for User Story 2

- [X] T020 [P] [US2] Write unit test for `RankingUseCase.GetProductRanking` in `internal/usecase/ranking_test.go` with a stub repository — verify it returns stores ordered by average_price ascending, ties broken by store_name ascending
- [X] T021 [P] [US2] Write unit test for currency isolation in `internal/usecase/ranking_test.go` — verify USD and Bs. entries are in separate currency groups
- [X] T022 [P] [US2] Write unit test for empty ranking in `internal/usecase/ranking_test.go` — verify a product with no aggregates returns an empty map, not an error
- [X] T023 [P] [US2] Write integration test for `GET /api/v1/products/{id}/prices` in `internal/adapter/http/ranking_handler_test.go` using `httptest` — verify 200 response shape with `currency_rankings` map (depends on T027)
- [X] T024 [P] [US2] Write integration test for error cases in `internal/adapter/http/ranking_handler_test.go` — verify 400 for invalid UUID, 401 for missing auth, 404 for non-existent product (depends on T027)

### Implementation for User Story 2

- [X] T025 [US2] Implement `GetProductRanking` method in `internal/adapter/postgres/price_aggregate_repository.go`: SELECT from `price_aggregates` JOIN `stores` WHERE product_id = $1 AND sample_count > 0 ORDER BY average_price ASC, store_name ASC (depends on T009)
- [X] T026 [US2] Create `RankingUseCase` in `internal/usecase/ranking.go` with method `GetProductRanking(ctx, productID string) (*ProductRankingResponse, error)` that calls `repo.GetProductRanking` and `repo.GetProductName`, groups results per currency, and builds the response (depends on T025, T010)
- [X] T027 [US2] Replace stub `RankingHandler` in `internal/adapter/http/ranking_handler.go` with real struct that accepts `*RankingUseCase` — implement `ProductPrices(w http.ResponseWriter, r *http.Request)` handler for `GET /api/v1/products/{id}/prices` with UUID validation, auth check, and JSON response per contract (depends on T026)
- [X] T028 [US2] Add route `authed.Get("/products/{id}/prices", rankingHandler.ProductPrices)` in `internal/adapter/http/router.go` (depends on T027)
- [X] T029 [US2] Update `cmd/api/main.go` to construct `RankingUseCase` with `PriceAggregateRepository` and pass it to `NewRankingHandler` (depends on T026, T027)
- [X] T030 [US2] Update `internal/adapter/http/router_test_helpers_test.go` to accept the new `RankingHandler` constructor signature (depends on T027)

**Checkpoint**: User Stories 1 AND 2 work independently. Confirm a receipt (US1), then query the per-product ranking endpoint (US2) to see fresh data.

---

## Phase 5: User Story 3 — Search Products by Name and Get Cheapest Store Per Currency (Priority: P1)

**Goal**: Expose `GET /api/v1/search?q=...` that searches products by normalized name and returns the cheapest store per currency for each match. Also replace the existing stub `rankingHandler.Search` with a real implementation.

**Independent Test**: Seed products "Arroz Blanco" and "Arroz Integral" with aggregates. Search for "arroz". Verify both products appear, each with the cheapest store per currency.

### Tests for User Story 3

- [X] T031 [P] [US3] Write unit test for `RankingUseCase.SearchProducts` in `internal/usecase/ranking_test.go` with a stub repository — verify it returns matching products with cheapest store per currency
- [X] T032 [P] [US3] Write unit test for empty search results in `internal/usecase/ranking_test.go` — verify no matches returns empty list, not error
- [X] T033 [P] [US3] Write unit test for short query rejection in `internal/usecase/ranking_test.go` — verify query < 3 characters returns an error
- [X] T034 [P] [US3] Write integration test for `GET /api/v1/search` in `internal/adapter/http/ranking_handler_test.go` — verify 200 response with `results` array, 400 for missing/short `q`, 401 for missing auth (depends on T037)

### Implementation for User Story 3

- [X] T035 [US3] Implement `SearchProducts` method in `internal/adapter/postgres/price_aggregate_repository.go`: SELECT matching products via `ILIKE` on `unaccent(canonical_name)` with `unaccent($1)` query, then for each product SELECT the cheapest store per currency from `price_aggregates` (depends on T025)
- [X] T036 [US3] Add `SearchProducts` method to `RankingUseCase` in `internal/usecase/ranking.go` with query validation (min 3 chars) and result assembly (depends on T035)
- [X] T037 [US3] Implement `Search` handler in `internal/adapter/http/ranking_handler.go` for `GET /api/v1/search?q=...` — validate `q` present and >= 3 chars, call `RankingUseCase.SearchProducts`, return JSON per `product-search-api-contract.md` (depends on T036)
- [X] T038 [US3] Update existing `Search` handler in `internal/adapter/http/ranking_handler.go` for `GET /api/v1/ranking/products/search` to call `RankingUseCase.SearchProducts` and return the legacy contract shape (flat `stores` array per product) to maintain Flutter app compatibility (depends on T036)
- [X] T039 [US3] Add route `authed.Get("/search", rankingHandler.Search)` in `internal/adapter/http/router.go` alongside the existing `ranking/products/search` route (depends on T037, T038)

**Checkpoint**: User Stories 1, 2, AND 3 work independently. Search for a product, see cheapest stores per currency.

---

## Phase 6: User Story 4 — Observation Freshness and Age Weighting (Priority: P2)

**Goal**: Apply a configurable age-threshold hard filter so that observations older than `PRICE_AGE_THRESHOLD_DAYS` are excluded from aggregate computation and ranking queries.

**Independent Test**: Seed two observations for the same triple: one 120 days ago at $0.50 and one yesterday at $2.00. With a 90-day threshold, verify only the $2.00 observation is included. Change threshold to 180 days, recompute, verify both are included.

### Tests for User Story 4

- [X] T040 [P] [US4] Write unit test in `internal/usecase/price_aggregate_recompute_test.go` — verify `RecomputeAggregate` is called with the configured age threshold value from config
- [X] T041 [P] [US4] Write integration test in `internal/adapter/postgres/price_aggregate_repository_test.go` — seed an old observation (120 days ago) and a fresh one (1 day ago), call `RecomputeAggregate` with threshold 90, verify `sample_count = 1` and average reflects only the fresh observation
- [X] T042 [P] [US4] Write unit test in `internal/usecase/ranking_test.go` — verify `GetProductRanking` excludes stores with `sample_count = 0` (stale stores don't appear)

### Implementation for User Story 4

- [X] T043 [US4] Add age-threshold WHERE clause to `RecomputeAggregate` in `internal/adapter/postgres/price_aggregate_repository.go`: `AND observed_at >= NOW() - ($thresholdDays || ' days')::interval` (depends on T009)
- [X] T044 [US4] Add `sample_count > 0` filter to `GetProductRanking` query in `internal/adapter/postgres/price_aggregate_repository.go` so stale aggregates (all observations aged out) are excluded from ranking (depends on T025)
- [X] T045 [US4] Pass `cfg.PriceAgeThresholdDays` from `cmd/api/main.go` to `ReceiptConfirmUseCase` and through to `PriceAggregateRecomputeUseCase.Execute` (depends on T003, T017, T043)
- [X] T046 [US4] Write a one-shot recompute function `RecomputeAll(ctx, ageThresholdDays int) error` in `internal/adapter/postgres/price_aggregate_repository.go` that recomputes all aggregates from scratch — used when the threshold changes (depends on T043)

**Checkpoint**: User Story 4 is functional — stale observations are filtered. Change `PRICE_AGE_THRESHOLD_DAYS` and recompute to verify.

---

## Phase 7: User Story 5 — Optional Proximity-Based Ordering with PostGIS (Priority: P3)

**Goal**: When the user provides `lat`, `long`, and optionally `radius_km`, the per-product ranking endpoint orders stores by proximity or filters by radius using PostGIS.

**Independent Test**: Seed two stores with coordinates (S1 near user, S2 far). Query the per-product ranking with user location and radius. Verify S1 appears before S2, and S2 is excluded when radius is small.

### Tests for User Story 5

- [X] T047 [P] [US5] Write unit test for `RankingUseCase.GetProductRanking` with proximity options in `internal/usecase/ranking_test.go` — verify `RankingQueryOptions{Lat, Long}` is passed through to the repository
- [X] T048 [P] [US5] Write integration test for proximity in `internal/adapter/http/ranking_handler_test.go` — verify stores without coordinates are included but sorted last, and stores beyond radius are excluded (depends on T051)

### Implementation for User Story 5

- [X] T049 [P] [US5] Update `stores` table seed data in test setup to include `lat`/`long` columns for integration tests in `internal/adapter/http/ranking_handler_test.go`
- [X] T050 [US5] Implement `GetProductRanking` with `RankingQueryOptions` proximity in `internal/adapter/postgres/price_aggregate_repository.go`: when Lat/Long provided, compute `ST_Distance(geo, ST_MakePoint(long, lat)::geography)` as `distance_km`; when RadiusKm provided, filter with `ST_DWithin(geo, ST_MakePoint(long, lat)::geography, radius_km * 1000)`; stores with NULL geo get `distance_km = NULL` and sort last (depends on T025)
- [X] T051 [US5] Update `RankingUseCase.GetProductRanking` in `internal/usecase/ranking.go` to accept optional `lat`, `long`, `radius_km` from query params and pass them as `RankingQueryOptions` to the repository (depends on T026, T050)
- [X] T052 [US5] Parse `lat`, `long`, `radius_km` query parameters in `ProductPrices` handler in `internal/adapter/http/ranking_handler.go` and pass to the use case (depends on T027, T051)
- [X] T053 [US5] Include `distance_km` field in the ranking response JSON in `internal/adapter/http/ranking_handler.go` — null when no location provided, number when proximity is active (depends on T052)

**Checkpoint**: All user stories complete. Proximity is opt-in and does not affect default ranking behavior.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Integration verification, contract compatibility, and final validation

- [X] T054 [P] Write integration test for existing `GET /api/v1/ranking/products/search` contract compatibility in `internal/adapter/http/ranking_handler_test.go` — verify the legacy endpoint still returns the shape from `specs/004-flutter-mobile-app/contracts/ranking-api-contract.md` with `results[].stores` flat array
- [X] T055 [P] Write integration test for currency isolation across endpoints in `internal/adapter/http/ranking_handler_test.go` — seed observations in both USD and Bs. for the same product, verify per-product ranking and search both return separate currency groupings
- [X] T056 [P] Write integration test for tie-breaking determinism in `internal/adapter/http/ranking_handler_test.go` — seed two stores with the same average_price, verify they are ordered by store_name ascending consistently across two requests
- [X] T057 Run all Go tests: `go test ./internal/...` and verify zero failures
- [X] T058 Run `docker compose up -d --build api` and execute quickstart.md scenarios 1-6 end-to-end
- [X] T059 [P] Verify `flutter analyze` on `mobile/` still passes (no Flutter changes required for this feature, but the existing ranking API client should still compile against the preserved contract)
- [X] T060 [P] Update `internal/adapter/http/router.go` test helpers to cover both new routes (`/products/{id}/prices` and `/search`) in `internal/adapter/http/router_test_helpers_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on T001 (migration) — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational (T004-T011) — no dependency on other stories
- **US2 (Phase 4)**: Depends on Foundational — reads from `price_aggregates` populated by US1, but can be tested independently by seeding SQL directly
- **US3 (Phase 5)**: Depends on Foundational + US2 (shares `RankingUseCase` and handler file) — the `Search` handler in `ranking_handler.go` is updated in this phase
- **US4 (Phase 6)**: Depends on US1 (modifies `RecomputeAggregate`) and US2 (modifies `GetProductRanking` query)
- **US5 (Phase 7)**: Depends on US2 (modifies `GetProductRanking` query and handler)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Foundational only — no story dependencies
- **US2 (P1)**: Foundational only — independently testable by seeding SQL directly into `price_aggregates`
- **US3 (P1)**: Foundational + shares `RankingUseCase` with US2 — best implemented after US2 to avoid merge conflicts in `ranking.go` and `ranking_handler.go`
- **US4 (P2)**: US1 (recompute) + US2 (ranking query) — modifies both
- **US5 (P3)**: US2 (proximity extends the ranking query) — modifies the repository and handler

### Within Each User Story

- Tests are written FIRST (Constitution Article III.3: TDD where practical)
- Entity/ports before use cases
- Use cases before handlers
- Repository implementation before handler integration
- Wiring in `main.go` last

### Parallel Opportunities

- T002, T003 can run in parallel with T001 (independent files)
- T004, T005, T008 can run in parallel (entity files, no interdependency)
- T009, T010 can run in parallel after T004, T006 (different methods in same file — coordinate)
- T012, T013, T014 (US1 tests) can run in parallel
- T020, T021, T022 (US2 tests) can run in parallel
- T031, T032, T033 (US3 tests) can run in parallel
- T040, T041, T042 (US4 tests) can run in parallel
- T047, T048 (US5 tests) can run in parallel
- T054, T055, T056, T059, T060 (polish) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Unit test for RecomputeUseCase basic flow in internal/usecase/price_aggregate_recompute_test.go"
Task: "Unit test for currency isolation in internal/usecase/price_aggregate_recompute_test.go"
Task: "Unit test for new product in internal/usecase/price_aggregate_recompute_test.go"

# Then implement in order:
Task: "Create PriceAggregateRecomputeUseCase in internal/usecase/price_aggregate_recompute.go"
Task: "Wire recompute into ReceiptConfirmUseCase in internal/usecase/receipt_confirm.go"
Task: "Update main.go wiring in cmd/api/main.go"
```

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task: "Unit test for ordering in internal/usecase/ranking_test.go"
Task: "Unit test for currency isolation in internal/usecase/ranking_test.go"
Task: "Unit test for empty ranking in internal/usecase/ranking_test.go"
Task: "Integration test for endpoint in internal/adapter/http/ranking_handler_test.go"
Task: "Integration test for errors in internal/adapter/http/ranking_handler_test.go"

# Then implement in order:
Task: "GetProductRanking in internal/adapter/postgres/price_aggregate_repository.go"
Task: "RankingUseCase in internal/usecase/ranking.go"
Task: "Handler in internal/adapter/http/ranking_handler.go"
Task: "Route in internal/adapter/http/router.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (migration + config)
2. Complete Phase 2: Foundational (entity, port, repository)
3. Complete Phase 3: US1 (recompute on confirm) → Test: confirm a receipt, verify `price_aggregates` has fresh rows
4. Complete Phase 4: US2 (per-product ranking endpoint) → Test: query `GET /api/v1/products/{id}/prices`
5. **STOP and VALIDATE**: Confirm a receipt, then query the ranking. The core "where to buy cheaper" loop is demonstrable.

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Recompute works → Demo: confirm receipt, inspect aggregates
3. Add US2 → Ranking endpoint works → Demo: query cheapest stores for a product
4. Add US3 → Search works → Demo: search "arroz", see cheapest stores per currency
5. Add US4 → Freshness filter → Demo: old prices excluded from ranking
6. Add US5 → Proximity → Demo: nearby stores first (optional)
7. Polish → Integration tests, contract compatibility, quickstart validation

### Parallel Team Strategy

With multiple developers:
1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (recompute use case + wiring)
   - Developer B: US2 (ranking use case + endpoint) — seeds SQL directly for testing
3. After US1 + US2 merge:
   - Developer A: US3 (search)
   - Developer B: US4 (age threshold)
4. After US3 + US4:
   - One developer: US5 (proximity)
   - Another: Polish (integration tests, contract compat)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Tests written BEFORE implementation per Constitution Article III.3
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- The existing Flutter ranking API client (`mobile/lib/features/ranking/`) does NOT need changes — the existing `ranking/products/search` contract is preserved and now returns real data instead of empty results