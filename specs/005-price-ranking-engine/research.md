# Research: Average-Price Engine and "Where to Buy Cheaper" Ranking

**Feature**: `005-price-ranking-engine`
**Date**: 2025-06-29

---

## R1 — Materialized aggregates vs. compute-on-query

### Decision
Compute aggregates on query (at read time) from `price_observations`,
with an optional `price_aggregates` table as a precomputed cache that is
updated synchronously on receipt confirmation.

### Rationale
- The MVP scale is small (single user demo, tens of receipts). A full
  table scan of `price_observations` filtered by `(product_id, store_id,
  currency)` with the existing index
  `idx_price_observations_product_store_date` returns in milliseconds.
- A materialized aggregate table adds write-path complexity (upsert on
  every confirmation) but saves read time only at scale (>10k
  observations per product).
- The project values simplicity (constitution Article VI.1: YAGNI).

### Alternatives considered
1. **Pure compute-on-query (no aggregate table).**
   Simpler. Every ranking query scans `price_observations` and aggregates
   on the fly. Acceptable for MVP scale; may degrade when observations
   exceed ~100k rows. Rejected as the only strategy because the spec
   explicitly introduces a `PriceAggregate` entity and the user
   description says "recompute the PriceAggregate" on confirmation.
2. **Materialized aggregates only (no query-time fallback).**
   Faster reads. Requires upsert logic on every confirmation and a
   strategy to handle threshold changes (stale rows must be purged).
   Rejected alone because it complicates the confirmation path and
   threshold changes require a full recompute pass.
3. **Hybrid (chosen).**
   Store `price_aggregates` as a precomputed cache updated synchronously
   on confirmation. Ranking queries read from the aggregate table. When
   the age threshold changes, a one-shot recompute function rebuilds the
   table. Query-time fallback scans observations only if the aggregate
   table is empty for a given (product, store, currency) triple — this
   handles the edge case of an existing product that has observations
   but no aggregate row yet (e.g., migration rollover).

---

## R2 — Age threshold: filter vs. weight

### Decision
Hard filter: observations older than the configurable threshold are
excluded from the aggregate computation.

### Rationale
- The spec accepts either filter or weighting ("either exclude
  observations older than a configurable threshold or weight them by
  age"). Filtering is simpler and the constitution favors simplicity
  (Article VI.1).
- In a high-inflation economy, old prices are not just "less relevant" —
  they are actively wrong. A hard cutoff is easier to reason about and
  to test.
- Weighted averages require defining a decay function (linear,
  exponential) and a half-life, which are speculative choices without
  domain data.

### Alternatives considered
1. **Exponential decay weighting.**
   Every observation contributes with `weight = exp(-age / half_life)`.
   Smooth, no hard cutoff. Rejected because choosing a half-life is
   speculative and the distinction between "old" and "very old" is
   unclear in a hyperinflation context where a month-old price can be
   10x off.
2. **Linear decay.**
   `weight = max(0, 1 - age / threshold)`. Simpler than exponential but
   still introduces a gradual fade that is hard to justify without
   domain calibration.
3. **Hard filter (chosen).**
   `WHERE observed_at >= NOW() - threshold`. Pruning is explicit,
   testable, and easy to reason about. Threshold is configurable via
   environment variable `PRICE_AGE_THRESHOLD_DAYS`.

---

## R3 — Aggregate recomputation strategy on confirmation

### Decision
Synchronous recompute within the same database transaction as the
receipt confirmation. After inserting `price_observations`, run an
`UPSERT` into `price_aggregates` for each affected
`(product_id, store_id, currency)` triple, recomputing average, min,
and sample count from the fresh observations within the threshold.

### Rationale
- The spec requires that the "very next query sees fresh data"
  (User Story 1) and that recomputation is synchronous (Assumptions:
  "eventual consistency or a delayed batch recompute is out of MVP
  scope").
- Doing it in the same transaction guarantees atomicity: either the
  confirmation and the aggregate update commit together, or both roll
  back. No partial state.
- The number of affected triples per receipt is small (one per line
  item, typically <20). The recompute query is a single
  `INSERT ... ON CONFLICT ... DO UPDATE` per triple or a batched CTE —
  cheap at MVP scale.

### Alternatives considered
1. **Async recomputation via a Redis queue (like the OCR worker).**
   Decouples confirmation from aggregation. Rejected because the spec
   demands synchronous freshness and adds queue complexity for no MVP
   benefit.
2. **Trigger-based recomputation (Postgres trigger on
   `price_observations` insert).**
   Transparent to the application. Rejected because it hides business
   logic in the database layer (violates Clean Architecture spirit —
   the domain should own the aggregation rule, not the DB), and makes
   the age-threshold configuration harder to surface in the application
   config.

---

## R4 — Accent-insensitive and case-insensitive search

### Decision
Use PostgreSQL `ILIKE` with `unaccent()` from the `unaccent` extension
for accent-insensitive, case-insensitive substring search on
`products.canonical_name`.

### Rationale
- The `products` table already stores `canonical_name` as
  lowercased whitespace-collapsed text (existing `normalizeName` in
  `receipt_repository.go`).
- `ILIKE` gives case-insensitive matching. `unaccent()` strips accents
  so "arroz" matches "Arroz" and "ARROZ".
- The `unaccent` extension ships with PostgreSQL and is enabled with
  `CREATE EXTENSION IF NOT EXISTS unaccent`.
- This is a query-time concern, not a normalization pipeline change
  (Assumptions: "accent-insensitive search is achieved through database
  collation or query-time normalization").

### Alternatives considered
1. **Full-text search (`tsvector` / `to_tsquery`).**
   Powerful for ranking by relevance. Overkill for MVP substring
   matching and adds a `tsvector` column + trigger maintenance.
2. **Trigram similarity (`pg_trgm`).**
   Good for fuzzy matching ("arroz" ≈ "arros"). Adds an extension and
   GIN index. Could be a future enhancement but not needed for exact
   substring search at MVP scale.
3. **Pre-computed normalized column with a second index.**
   Adds a column storing an accent-stripped version of `canonical_name`.
   Rejected in favor of `unaccent()` at query time to avoid storing
   redundant data.

---

## R5 — Currency isolation (Constitution Article V)

### Decision
Every aggregate and ranking query is scoped by `currency` as a
_grouping key_. The `price_aggregates` table has a composite key
`(product_id, store_id, currency)`. No query ever computes an average
across currencies.

### Rationale
- Constitution Article V.1: "Averages are computed per currency;
  currencies are NEVER mixed within a single average."
- The existing `price_observations` table already has a `currency`
  column (NOT NULL).
- The spec's FR-002 and SC-003 require per-currency grouping in the
  ranking response.
- The response schema groups stores per currency (a map keyed by
  currency) rather than flattening.

### Alternatives considered
1. **Separate table per currency.**
   Rejected — extreme complexity for no benefit. The `currency` column
   is the natural partition key.

---

## R6 — Proximity ordering with PostGIS

### Decision
Add `lat`/`long` columns to the `stores` table (already planned by the
constitution Article V.2 and the existing migration). Implement
proximity as an OPTIONAL query parameter in the ranking endpoint:
`lat`, `long`, and `radius_km`. When provided, use
`ST_DWithin` and `ST_Distance` from PostGIS to filter and order stores.

### Rationale
- PostGIS is already installed and active in the Docker Postgres
  container (confirmed by the `tiger` and `topology` schemas present in
  the database).
- The constitution mandates Store geolocation (Article V.2).
- The spec marks proximity as P3 (optional, deferred). The plan
  includes the schema change and a thin query path, but the mobile app
  does not need to send location in the MVP — the endpoint simply
  ignores the parameters when absent (FR-016).

### Alternatives considered
1. **Haversine formula in application code.**
   Rejected — PostGIS is already available and is more accurate and
   index-friendly (GiST index on `geography` column).
2. **Defer all proximity work to a future spec.**
   Rejected because the spec explicitly includes it as an optional
   acceptance criterion, and the schema change is trivial to include
   now. The query path is opt-in and does not affect the default
   ranking.

---

## R7 — Deterministic tie-breaking

### Decision
When two stores have the same average price for a product, break ties
by `store_name` ascending as a stable secondary sort key.

### Rationale
- The spec requires a deterministic ranking order (FR-010, SC-005).
- Store name is stable, human-readable, and already indexed in the
  `stores` table.
- Using a secondary sort on the store name is a standard practice for
  stable pagination.

### Alternatives considered
1. **Tie-break by `created_at` (oldest store first).**
   Not user-meaningful.
2. **Tie-break by `store_id` (UUID).**
   Stable but arbitrary and confusing in debugging.

---

## R8 — Endpoint design: new vs. existing contract alignment

### Decision
Implement two endpoints:
1. `GET /api/v1/products/{id}/prices` — per-product ranking (new
   endpoint, defined by this feature's spec).
2. `GET /api/v1/search?q=...` — product search returning cheapest
   store per currency (new endpoint, defined by this feature's spec).

The existing stub `GET /api/v1/ranking/products/search` (from spec 004)
is replaced with a real implementation that serves the same contract
shape but routes to the new search use case. The mobile app already
consumes this contract; no Flutter change is required for the ranking
to populate.

### Rationale
- The spec defines two new endpoint paths. The existing
  `ranking/products/search` path was a placeholder contract for the
  Flutter app; the real implementation must satisfy that contract.
- We honor constitution Article IV.1 (versioned contracts): all new
  endpoints live under `/api/v1/`.
- We honor Article IV.4 (Flutter consumes ONLY published contracts):
  the existing `ranking-api-contract.md` is preserved, and the new
  per-product endpoint gets its own contract document.

### Alternatives considered
1. **Replace `ranking/products/search` entirely with `/search`.**
   Rejected because it breaks the already-published contract that the
   Flutter app consumes. Instead, the existing path is kept and serves
   the real implementation; the new `/search` path is an alias.
2. **Deprecate `ranking/products/search` and only add `/search`.**
   Rejected for the same breaking-contract reason.