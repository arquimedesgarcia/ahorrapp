# Phase 1 — Data Model

**Feature**: 006-loyalty-points
**Date**: 2026-06-29

This feature extends the existing data model rather than rewriting it. New
SQL elements are confined to migration `000005`.

## Entities

### `users` (existing, extended use)

No schema change. Reuses the existing `points INTEGER NOT NULL DEFAULT 0`
column as a denormalized cache of `SUM(loyalty_transactions.points)` for
the user.

**Invariants**:
- `users.points == SUM(loyalty_transactions.points WHERE user_id = users.id)`.
- Updated atomically with each insert in the same transaction (R-08).

### `receipts` (existing, unchanged schema)

Used read-only by the award use case (status must be `CONFIRMED` at award
time). The `id` becomes the idempotency key for awarding.

### `loyalty_transactions` (existing, extended)

Existing columns (migration `000003`):

| Column      | Type        | Notes                                   |
|-------------|-------------|-----------------------------------------|
| id          | UUID PK     | `gen_random_uuid()`                      |
| user_id     | UUID FK     | `REFERENCES users(id) ON DELETE CASCADE`|
| points      | INTEGER     | not null; can be 0 or negative later    |
| reason      | TEXT        | not null; machine-readable reason       |
| created_at  | TIMESTAMPTZ | not null; default `NOW()`               |

**New column and constraints (migration `000005`)**:

| Change                                                                                       | Why |
|----------------------------------------------------------------------------------------------|-----|
| `ALTER TABLE loyalty_transactions ADD COLUMN receipt_id UUID REFERENCES receipts(id) ON DELETE SET NULL` | Ties award rows to the receipt that triggered them (FR-002); nullable to leave room for future non-receipt movements (R-08). |
| `CREATE UNIQUE INDEX uniq_loyalty_tx_receipt ON loyalty_transactions(receipt_id) WHERE receipt_id IS NOT NULL` | Enforces "award once" at the data layer (FR-004). Partial unique index permits future null-receipt rows to coexist. |
| `CREATE INDEX idx_loyalty_tx_user_day ON loyalty_transactions(user_id, created_at)` | Supports the daily-cap query (FR-005). |

### `loyalty_configuration` (NOT a table)

Per R-05 configuration is delivered via environment variables read by
`internal/config/config.go` and injected into the use case. There is no
`loyalty_configuration` table. Listed here only to call out the decision
explicitly.

### Key configuration values (process env, not SQL)

| Env var                                | Default | Meaning                                |
|----------------------------------------|---------|----------------------------------------|
| `LOYALTY_BASE_POINTS`                  | 10      | Points per confirmed receipt (FR-001)  |
| `LOYALTY_FIRST_OBSERVATION_BONUS`      | 5       | Bonus for a brand-new (product, store) or store |
| `LOYALTY_DATA_COMPLETION_BONUS`        | 3       | Bonus for filling optional fields      |
| `LOYALTY_DAILY_AWARD_CAP`              | 20      | Max point-granting receipts per user per UTC day |

## Reason codes (enumeration, not stored in DB)

Documented here as the canonical list so the use case and tests agree.
Stored verbatim in `loyalty_transactions.reason`.

| Reason                       | Points        | When                                                          |
|------------------------------|---------------|---------------------------------------------------------------|
| `receipt_confirmed`          | base          | New receipt confirmed, within daily cap                        |
| `first_observation_product`  | bonus         | First `price_observations` row for a (product, store) pair     |
| `first_observation_store`    | bonus         | A new store row was created by this confirmation               |
| `data_completion`            | bonus         | Optional fields all populated by the user                      |
| `daily_limit_reached`        | 0             | Confirmation beyond the daily cap (FR-005)                     |

Every award for a single receipt produces **one** summary row whose
`points` is the sum of base + earned bonuses and whose `reason` is a
`;`-joined list of the contributing reason codes (e.g.
`receipt_confirmed;first_observation_product;data_completion`). This
keeps the invariant "one receipt_id → one award row" trivially true and
the history compact. Tests assert on the reason string content.

## Validation rules from spec

These are enforced by the use case and/or DB constraints:

- Only `CONFIRMED` receipts trigger awarding (FR-012). Enforced by the
  confirm flow being the only caller of the award use case.
- Same receipt never grants points twice (FR-004). Enforced by the
  unique partial index above.
- Daily cap counts only positive point-granting receipts (FR-005). The
  award use case counts distinct `receipt_id` rows with `points > 0`
  for the user on the current UTC day.
- `users.points == SUM(movements)` (FR-011). Atomic within the award
  transaction; integration test verifies equality after several awards.

## State transitions

Not applicable to `loyalty_transactions` itself (append-only). The
`receipt.status` transition that matters — `NEEDS_REVIEW → CONFIRMED`
— already exists in `ReceiptRepository.ConfirmReceipt` and is unchanged.

## Backfill / migration safety

- `ALTER TABLE` adds a nullable column; existing rows receive NULL
  `receipt_id` (no data loss; already-awarded points, if any, remain
  valid).
- The unique partial index is created `CONcurrently`-friendly for the
  MVP scale (single user in dev). Field test is in quickstart.md.
- Down migration drops the column and the two indexes.