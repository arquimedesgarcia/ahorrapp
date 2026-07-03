# Data Model: Average-Price Engine and Ranking

**Feature**: `005-price-ranking-engine`
**Date**: 2025-06-29

---

## New Entities

### PriceAggregate

The precomputed aggregate cache for a (product, store, currency) triple.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `product_id` | UUID | NOT NULL, FK → products(id) | Canonical product |
| `store_id` | UUID | NOT NULL, FK → stores(id) | Store / merchant |
| `currency` | TEXT | NOT NULL, CHECK IN ('USD', 'Bs.') | Isolation key (Art. V) |
| `average_price` | NUMERIC(12,2) | NOT NULL | Mean of fresh observations |
| `min_price` | NUMERIC(12,2) | NOT NULL | Minimum of fresh observations |
| `sample_count` | INT | NOT NULL, CHECK >= 0 | Fresh observations included |
| `last_observed_at` | TIMESTAMPTZ | NOT NULL | Most recent observation included |
| `updated_at` | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | Last recompute timestamp |

**Primary key**: `(product_id, store_id, currency)` — composite,
guarantees one row per triple.

**Index**: `idx_price_aggregates_product_currency`
on `(product_id, currency)` for ranking lookups.

---

## Modified Entities

### stores (existing table — add geolocation)

| New Field | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `lat` | DOUBLE PRECISION | NULL | Store latitude (PostGIS-ready) |
| `long` | DOUBLE PRECISION | NULL | Store longitude |

Both are nullable because legacy stores created before this migration do
not have coordinates. The proximity query treats stores with NULL lat/long
as "unknown distance" — included without proximity ordering, sorted last
when proximity is active (per spec Assumptions).

**PostGIS column (optional, for proximity)**: A generated
`geography(Point, 4326)` column `geo` derived from `(lat, long)` with a
GiST index, enabling `ST_DWithin` and `ST_Distance`. This column is
populated only when both lat and long are non-NULL.

---

## Existing Entities (unchanged, consumed)

### price_observations (existing)

Source of truth for all aggregates. Already has:
- `product_id`, `store_id`, `currency`, `unit_price`, `observed_at`,
  `receipt_id`.
- Index `idx_price_observations_product_store_date` on
  `(product_id, store_id, observed_at DESC)`.

The age-threshold filter (`observed_at >= NOW() - threshold`) uses this
index efficiently.

### products (existing)

Used for search (`canonical_name`) and as the FK target for
`price_aggregates.product_id`.

### stores (existing, modified)

Used as FK target and for store name / branch in ranking output.

---

## Migration Script

**File**: `migrations/000004_price_aggregates_store_geo.up.sql`

```sql
-- 1. Add geolocation to stores
ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS long DOUBLE PRECISION;

-- 2. PostGIS geography column + GiST index for proximity queries
ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS geo geography(Point, 4326)
    GENERATED ALWAYS AS (
        CASE WHEN lat IS NOT NULL AND long IS NOT NULL
             THEN ST_MakePoint(long, lat)::geography
        END
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_stores_geo ON stores USING GIST (geo);

-- 3. Price aggregates cache table
CREATE TABLE IF NOT EXISTS price_aggregates (
    product_id       UUID        NOT NULL REFERENCES products(id),
    store_id         UUID        NOT NULL REFERENCES stores(id),
    currency         TEXT        NOT NULL CHECK (currency IN ('USD', 'Bs.')),
    average_price    NUMERIC(12,2) NOT NULL,
    min_price        NUMERIC(12,2) NOT NULL,
    sample_count     INT         NOT NULL CHECK (sample_count >= 0),
    last_observed_at TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, store_id, currency)
);

CREATE INDEX IF NOT EXISTS idx_price_aggregates_product_currency
    ON price_aggregates(product_id, currency);

-- 4. Accent-insensitive search support
CREATE EXTENSION IF NOT EXISTS unaccent;
```

**File**: `migrations/000004_price_aggregates_store_geo.down.sql`

```sql
DROP INDEX IF EXISTS idx_price_aggregates_product_currency;
DROP TABLE IF EXISTS price_aggregates;
DROP INDEX IF EXISTS idx_stores_geo;
ALTER TABLE stores DROP COLUMN IF EXISTS geo;
ALTER TABLE stores DROP COLUMN IF EXISTS long;
ALTER TABLE stores DROP COLUMN IF EXISTS lat;
```

---

## State Transitions

### PriceAggregate recomputation flow

```
Receipt confirmed
    │
    ▼
Insert price_observations (existing)
    │
    ▼
For each (product_id, store_id, currency) triple:
    │
    ▼
    SELECT AVG(unit_price), MIN(unit_price), COUNT(*), MAX(observed_at)
    FROM price_observations
    WHERE product_id = $1
      AND store_id = $2
      AND currency = $3
      AND observed_at >= NOW() - ($threshold || ' days')::interval
    │
    ▼
    INSERT INTO price_aggregates (...)
    ON CONFLICT (product_id, store_id, currency)
    DO UPDATE SET average_price = EXCLUDED.average_price,
                  min_price      = EXCLUDED.min_price,
                  sample_count   = EXCLUDED.sample_count,
                  last_observed_at = EXCLUDED.last_observed_at,
                  updated_at     = NOW();
```

### Stale aggregate handling

When all observations for a triple are older than the threshold:
- The SELECT returns 0 rows.
- The UPSERT writes `sample_count = 0`, `average_price = 0`,
  `min_price = 0`, `last_observed_at` = (the latest stale observation's
  timestamp, or null if none).
- Ranking queries filter out rows with `sample_count = 0`, so stale
  stores do not appear in the ranking (FR-009).

---

## Validation Rules (from spec requirements)

| Rule | Source | Enforcement |
|------|--------|-------------|
| Currency is mandatory on every observation | FR-002, Art. V.1 | `price_observations.currency NOT NULL` (existing) + `price_aggregates.currency NOT NULL` (new) + CHECK constraint |
| Age threshold configurable | FR-004 | `PRICE_AGE_THRESHOLD_DAYS` env var, read in `config.Load()` |
| Product ID must be valid UUID | FR-014 | chi URL param + UUID parse validation in handler |
| Search query minimum length | SC-007 | Handler rejects `q` with < 3 characters with HTTP 400 |
| Ranking ordered cheapest first | FR-005/016 | `ORDER BY average_price ASC, store_name ASC` |
| No currency mixing | FR-002, SC-003 | All queries GROUP BY / partition on `currency` |