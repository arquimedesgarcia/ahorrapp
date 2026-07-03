# Quickstart — Validate the Loyalty Feature End-to-End

**Feature**: 006-loyalty-points
**Date**: 2026-06-29

This is a runnable validation guide, not a full test suite. Implementation
code is in `tasks.md`. Each scenario below maps to one or more acceptance
criteria in `spec.md` and is reproducible with the local Docker stack and
the project's test commands.

## Prerequisites

1. Local stack up (no cloud, Art. VI):
   ```powershell
   docker compose up -d postgres redis minio ocr
   ```
2. Go toolchain available (`go version` prints 1.22+).
3. The backend runs migrations on startup (see `cmd/api/main.go`); no
   manual `migrate up` is required. To reset data:
   ```powershell
   docker compose down -v
   docker compose up -d postgres redis minio ocr
   ```

## Scenario A — Confirmed receipt earns points exactly once (FR-001, FR-004)

1. Register a user and obtain a JWT:
   ```bash
   TOKEN=$(curl -s -XPOST localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"a@example.com","password":"pw123456","display_name":"A"}' \
     | jq -r .token)
   ```
2. Upload a receipt, wait until its status is `NEEDS_REVIEW`, fetch the
   editable summary, and POST the confirmation to
   `/api/v1/receipts/{id}/confirm` with a fully-populated payload.
3. Expected: `GET /api/v1/me/loyalty` returns `balance` equal to
   `LOYALTY_BASE_POINTS` (+bonuses if applicable) and one history entry
   whose reason contains `receipt_confirmed`.
4. Confirm the *same* receipt id again (or resubmit the same image and
   confirm the duplicate). Expected: no second movement, `balance`
   unchanged. (The confirm endpoint itself may return a conflict or a
   no-op; what matters is the loyalty history does not grow.)

Run the unit test:
```powershell
go test ./internal/usecase -run TestLoyaltyAward
```

## Scenario B — Daily cap allows confirmation but grants no points (FR-005)

1. Set `LOYALTY_DAILY_AWARD_CAP=2` for the API process and restart it.
2. Confirm 3 distinct receipts for the same user in the same UTC day.
3. Expected: history contains 3 movements. The first two have positive
   `points` and reason `receipt_confirmed`. The third has
   `points = 0` and reason `daily_limit_reached`. `balance` equals the
   sum of the first two movements only.
4. The third receipt MUST still be `CONFIRMED` (verify via
   `GET /api/v1/receipts/{id}`) — the price engine was fed.

```powershell
go test ./internal/usecase -run TestLoyaltyAward_DailyLimit
```

## Scenario C — First-observation bonus (FR-006)

1. Confirm a receipt that introduces a brand-new `(product, store)`
   pair (no prior `price_observations` row).
2. Expected: history reason contains
   `first_observation_product` and `points` includes the
   `LOYALTY_FIRST_OBSERVATION_BONUS` delta.
3. Confirm a second receipt with the same `(product, store)` pair.
   Expected: no `first_observation_product` in the reason; only base
   points awarded.

```powershell
go test ./internal/usecase -run TestLoyaltyAward_FirstObservation
```

## Scenario D — Data-completion bonus (FR-006)

1. Confirm a receipt where the user filled every optional field
   (`purchase_date`, `total`, and per-item `quantity`, `unit_price`,
   `currency`). Expected reason contains `data_completion`.
2. Confirm a receipt missing at least one optional field. Expected: no
   `data_completion` in the reason.

```powershell
go test ./internal/usecase -run TestLoyaltyAward_DataCompletion
```

## Scenario E — Balance and history are queryable and consistent (FR-003, FR-011, SC-003)

1. After confirming several receipts, call
   `GET /api/v1/me/loyalty`.
2. Expected: `balance` equals `SUM(history[].points)` and the history is
   ordered by `created_at DESC`.
3. An empty history case: a fresh user calls the endpoint. Expected:
   `{"balance":0,"history":[]}` with HTTP 200 (not an error).

```powershell
go test ./internal/usecase -run TestLoyaltyQuery
go test ./internal/adapter/http -run TestLoyaltyHandler
```

## Scenario F — Unauthenticated access is rejected (FR-010, SC-007)

```bash
curl -i localhost:8080/api/v1/me/loyalty
```
Expected: `401 Unauthorized` with
`{"error":"invalid or expired token"}`. No DB query is performed on
behalf of an anonymous caller (verifiable in the integration test by
pointing the use case at a fake repo and asserting it was never called).

```powershell
go test ./internal/adapter/http -run TestLoyaltyHandler_Unauthenticated
```

## Scenario G — Cross-user isolation (FR-009, SC-007)

1. Sign in as user A, confirm a receipt, observe A's balance.
2. Sign in as user B, call `/api/v1/me/loyalty`. Expected:
   `balance = 0` and an empty history — B never sees A's movements.

```powershell
go test ./internal/adapter/http -run TestLoyaltyHandler_CrossUser
```

## All-up check

```powershell
go test ./internal/usecase ./internal/adapter/http ./internal/adapter/postgres
go vet ./...
docker compose down -v
```

A clean exit code of `go test` plus `go vet` success indicates the
feature is functionally correct against this quickstart guide.