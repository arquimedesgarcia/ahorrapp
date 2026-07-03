# Quickstart: Average-Price Engine and Ranking

**Feature**: `005-price-ranking-engine`
**Date**: 2025-06-29

Validation scenarios that prove the feature works end-to-end on the local
Docker stack. No production deployment required.

## Prerequisites

- Docker and Docker Compose installed.
- The full stack running: `docker compose up -d` from the repo root.
- Migrations applied automatically on API container startup
  (migration `000004` is included in this feature).
- `curl` (or `Invoke-WebRequest` on PowerShell) and `jq` for JSON
  inspection.

## Setup

```powershell
# 1. Start the stack (or rebuild after code changes)
docker compose up -d --build api

# 2. Verify the new migration ran
docker exec ahorrapp-postgres psql -U ahorrapp -d ahorrapp -c "\d price_aggregates"
docker exec ahorrapp-postgres psql -U ahorrapp -d ahorrapp -c "\d stores"

# 3. Confirm the API is healthy
curl http://localhost:8080/api/v1/health
```

## Scenario 1 — Confirm a Receipt and Verify Aggregate Recomputation

This scenario proves User Story 1 (FR-001, FR-002, SC-001, SC-006).

### Steps

```powershell
# Register a test user and get a JWT
$resp = curl -s -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"aggregation@dev.local","password":"test1234","display_name":"Aggregator"}'
$token = ($resp | jq -r .token)

# Log in (if already registered)
$resp = curl -s -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"aggregation@dev.local","password":"test1234"}'
$token = ($resp | jq -r .token)

# Upload a receipt image (any small JPEG)
$upload = curl -s -X POST http://localhost:8080/api/v1/receipts `
  -H "Authorization: Bearer $token" `
  -F "image=@some-receipt.jpg"
$receiptId = ($upload | jq -r .receipt_id)

# Confirm the receipt with line items (price observations)
curl -s -X POST "http://localhost:8080/api/v1/receipts/$receiptId/confirm" `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "store": {"name": "TestStore"},
    "purchase_date": "2025-06-29",
    "total": 3.60,
    "items": [
      {"raw_text": "PAN", "quantity": 1, "unit_price": 1.20, "currency": "USD"},
      {"raw_text": "LECHE", "quantity": 2, "unit_price": 1.20, "currency": "USD"}
    ]
  }'

# Verify the price_aggregates table has fresh rows
docker exec ahorrapp-postgres psql -U ahorrapp -d ahorrapp -c `
  "SELECT product_id, store_id, currency, average_price, min_price, sample_count
   FROM price_aggregates ORDER BY updated_at DESC LIMIT 5;"
```

### Expected Outcome

- The confirm endpoint returns HTTP 204.
- `price_aggregates` has one row per (product, store, currency) triple
  from the receipt. `sample_count` matches the number of line items for
  that triple. `average_price` equals the mean of the unit prices.

## Scenario 2 — Per-Product Ranking Endpoint

This scenario proves User Story 2 (FR-005, FR-006, FR-009, FR-010,
FR-014, SC-003, SC-005).

### Steps

```powershell
# Get a product ID from the aggregates
$productId = docker exec ahorrapp-postgres psql -U ahorrapp -d ahorrapp -t -c `
  "SELECT product_id::text FROM price_aggregates LIMIT 1"
$productId = $productId.Trim()

# Query the per-product ranking
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/products/$productId/prices" | jq
```

### Expected Outcome

- HTTP 200 with a `currency_rankings` object keyed by currency.
- Stores within each currency are ordered by `average_price` ascending.
- Ties in average price are broken by `store_name` ascending.
- Each store entry includes `average_price`, `min_price`, `currency`,
  `sample_count`, and `last_observed_at`.
- A non-existent product ID returns HTTP 404.
- An invalid UUID returns HTTP 400.

## Scenario 3 — Product Search Endpoint

This scenario proves User Story 3 (FR-007, FR-008, FR-013, SC-002,
SC-007).

### Steps

```powershell
# Search for a product (assumes "PAN" was confirmed above)
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/search?q=pan" | jq

# Search with a short query (must be rejected)
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/search?q=pa"

# Search with no query (must be rejected)
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/search"
```

### Expected Outcome

- `q=pan` returns HTTP 200 with `results` containing at least one
  product whose normalized name matches "pan". Each result has a
  `best_prices` map keyed by currency with the cheapest store.
- `q=pa` returns HTTP 400 with `query must be at least 3 characters`.
- Missing `q` returns HTTP 400 with `query parameter 'q' is required`.
- Accent-insensitive: `q=arroz` matches "Arroz Blanco" and "Arroz
  Integral" (if they exist).

## Scenario 4 — Age Threshold Filtering

This scenario proves User Story 4 (FR-003, FR-004, FR-009, SC-004).

### Steps

```powershell
# Check the current threshold (default 90 days)
docker exec ahorrapp-api printenv PRICE_AGE_THRESHOLD_DAYS

# Seed an old observation directly (120 days ago)
docker exec ahorrapp-postgres psql -U ahorrapp -d ahorrapp -c "
INSERT INTO price_observations (product_id, store_id, unit_price, currency, observed_at, receipt_id)
SELECT p.id, s.id, 0.50, 'USD', NOW() - INTERVAL '120 days', r.id
FROM products p, stores s, receipts r
WHERE p.canonical_name = 'pan'
LIMIT 1;"

# Manually recompute aggregates for that product/store/currency
# (the recompute function will pick up the new stale observation and
#  filter it out because it is older than the 90-day threshold)
# ... recompute is triggered by the next receipt confirmation, or
# a one-shot recompute task (see tasks.md)

# Query the ranking again — the stale price should NOT affect the average
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/products/$productId/prices" | jq

# Change the threshold to 180 days and restart the API
docker compose stop api
$env:PRICE_AGE_THRESHOLD_DAYS = "180"
docker compose run --rm -e PRICE_AGE_THRESHOLD_DAYS=180 api

# Now the 120-day-old observation IS included, lowering the average
```

### Expected Outcome

- With a 90-day threshold, the 120-day-old observation is excluded
  and does not appear in the aggregate.
- With a 180-day threshold, the 120-day-old observation is included
  and the `average_price` decreases.
- Stores whose only observations are stale do not appear in the
  ranking (filtered out by `sample_count = 0`).

## Scenario 5 — Currency Isolation

This scenario proves FR-002, SC-003.

### Steps

```powershell
# Confirm a receipt with items in BOTH USD and Bs.
curl -s -X POST "http://localhost:8080/api/v1/receipts/$receiptIdBs/confirm" `
  -H "Authorization: Bearer $token" `
  -H "Content-Type: application/json" `
  -d '{
    "store": {"name": "BolivarStore"},
    "purchase_date": "2025-06-29",
    "total": 500.00,
    "items": [
      {"raw_text": "PAN", "quantity": 1, "unit_price": 125.00, "currency": "Bs."}
    ]
  }'

# Query the ranking for the product "PAN"
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/products/$productId/prices" | jq
```

### Expected Outcome

- The response has two keys in `currency_rankings`: `"USD"` and `"Bs."`.
- The USD average does not include the Bs. price and vice versa.
- No single average mixes currencies.

## Scenario 6 — Existing Flutter Contract Compatibility

This scenario proves that the existing `ranking/products/search`
endpoint still returns the expected shape (Article IV.4).

### Steps

```powershell
curl -s -H "Authorization: Bearer $token" `
  "http://localhost:8080/api/v1/ranking/products/search?q=pan" | jq
```

### Expected Outcome

- HTTP 200 with the shape documented in
  `specs/004-flutter-mobile-app/contracts/ranking-api-contract.md`.
- `results[].stores` is a flat array of stores with `average_price`,
  `currency`, and `sample_count`.
- No breaking change to the Flutter app.

## Running Tests

```powershell
# Run all Go tests
go test ./internal/...

# Run only the ranking-related tests
go test ./internal/usecase/ -run "Ranking|Aggregate"
go test ./internal/adapter/http/ -run "Ranking"
go test ./internal/adapter/postgres/ -run "Aggregate"

# Run with coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```