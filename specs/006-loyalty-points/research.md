# Phase 0 — Research & Decisions

**Feature**: 006-loyalty-points
**Date**: 2026-06-29

This feature has no NEEDS CLARIFICATION markers in the spec. Research tasks
focus on selecting the simplest project-aligned approach for each open
technical decision left to the plan level (per spec Assumptions).

---

## R-01 — Idempotency mechanism for "no double award"

**Decision**: Add a nullable `receipt_id` column with a UNIQUE constraint
to `loyalty_transactions`. Only the "base / bonus" award rows reference a
receipt; future non-receipt movements (redemptions) leave `receipt_id`
NULL. Awarding is performed inside the same Postgres transaction that
marks the receipt `CONFIRMED`, OR as a follow-up idempotent insert that
conflicts on `(user_id, receipt_id)`.

**Rationale**: The UNIQUE constraint makes the "award once" rule
un-cheatable at the data layer (FR-004) — even a buggy caller or a
retried request cannot insert a duplicate. It also enables the daily-cap
query to count `WHERE created_at >= start_of_day AND points > 0`
efficiently via the existing `(user_id, created_at DESC)` index.

**Alternatives considered**:
- A boolean `points_granted` flag on `receipts`: simpler but does not
  protect against duplicate award rows created by retries; would require
  application-level locking.
- Application-level check-then-insert: introduces a TOCTOU race under
  concurrent double-confirm attempts.

---

## R-02 — Daily limit semantics

**Decision**: The daily cap counts the number of *point-granting* receipts
in the current UTC day for that user — i.e., the number of distinct
`receipt_id` values with positive-sum points in `loyalty_transactions`
on the current UTC date. The cap is checked *before* awarding. When the cap
is reached, the receipt is still confirmed (status `CONFIRMED`, feeds the
price engine) but the award use case records a single
`LoyaltyTransaction` with `points = 0` and `reason = "daily_limit_reached"`
so the user's history explains why no points were granted.

**Rationale**: Advancing the price engine is the higher-order value;
points are the incentive layer. Zero-point movements make the history
explainable (FR-005) and keep the balance equal to the sum of movements
(FR-011). UTC keeps server-side reasoning simple; the spec explicitly
allows UTC as a reasonable default.

**Alternatives considered**:
- Reject the confirmation when the cap is reached: violates spec
  (receipt must still be confirmed and feed the price engine).
- Per-product/per-store cap instead of per-user: not in the spec scope.

---

## R-03 — Detecting "first observation"

**Decision**: A pair `(product_id, store_id)` is a "first observation" iff
no `price_observations` row exists for that pair *before* the current
confirmation transaction inserts the new ones. Detection is done in SQL
inside the same transaction: `SELECT EXISTS(SELECT 1 FROM price_observations WHERE product_id = $1 AND store_id = $2)` returns false for the
candidate pairs, *then* the new observations are inserted, *then* points
are computed. The "first store" bonus is granted if the store row was
created by this confirmation (i.e., the store did not previously exist),
detected by inspecting the store-creation operation already performed by
`ResolveOrCreateStore`.

**Rationale**: Performing detection inside the same DB transaction as the
insert naturally eliminates the race (R-04 edge case). It needs no extra
column on `price_observations`.

**Alternatives considered**:
- Application-level `FirstObservationChecker` reading *before* the
  transaction opens: introduces a TOCTOU race.
- Adding an `is_first` boolean to `price_observations`: redundant, and
  already derivable from absence of prior rows.

---

## R-04 — Detecting "data completion"

**Decision**: A receipt qualifies for the data-completion bonus when the
user filled in optional (`omitempty`) fields that the OCR pipeline usually
leaves null: `purchase_date`, `total`, and per-item `quantity` and
`currency`. Concretely: bonus granted iff `payload.PurchaseDate` is set,
`payload.Total > 0`, and *all* items have non-nil `quantity`, non-nil
`unit_price`, and non-nil `currency`. The same validation already done by
the confirm use case (`unit_price` and `currency` are required) means the
incremental condition is primarily `quantity` on every item plus
`purchase_date`/`total`.

**Rationale**: Clear, boolean, mechanical test (one SQL/Go predicate), no
new persisted flag required. Documented in `data-model.md`.

**Alternatives considered**:
- Persist an explicit `data_completed_at` flag: simpler to query but
  redundant with the existing payload validation.

---

## R-05 — Configuration delivery

**Decision**: Add `LOYALTY_BASE_POINTS` (default 10),
`LOYALTY_FIRST_OBSERVATION_BONUS` (default 5),
`LOYALTY_DATA_COMPLETION_BONUS` (default 3), and
`LOYALTY_DAILY_AWARD_CAP` (default 20) to `internal/config/config.go`,
parsed via `caarlos0/env/v11` exactly like the existing fields.
Configuration is read once at startup and passed via constructor
dependency injection to `LoyaltyAwardUseCase`. The plan does not
introduce a runtime-mutable configuration table (YAGNI per Art. VI.1).

**Rationale**: Mirrors the existing pattern (`PRICE_AGE_THRESHOLD_DAYS`).
No new dependency. Satisfies FR-008 (configurable without code changes).

**Alternatives considered**:
- A `loyalty_config` table loaded per-request: more flexible but heavier
  and not needed for the MVP.

---

## R-06 — Endpoint shape for `/api/v1/me/loyalty`

**Decision**: Add a new canonical `GET /api/v1/me/loyalty` returning
`{ "balance": int, "history": [ { id, points, reason, created_at, receipt_id? } ] }`.
History is ordered by `created_at DESC`, limited to the latest 100
movements (FR-003, SC-006). The history includes zero-point movements
(daily-limit-reached) so the user always understands why. Field
`receipt_id` is included when the movement is tied to a receipt, omitted
otherwise. The existing `GET /api/v1/users/me/points` is kept as a
backward-compatible alias returning the same JSON shape so the Flutter
app keeps working unchanged until a later feature migrates it.

**Rationale**: One canonical, well-documented endpoint; preserves the
existing contract (Art. IV). Limit of 100 keeps the response fast and
bounded; pagination is explicitly out of scope for this feature (Art.

VI.1).

**Alternatives considered**:
- Rename `/users/me/points` to `/me/loyalty`: breaking change (Art. IV
  forbids breaking under v1).
- Cursor pagination: over-engineering for the MVP user volume.

---

## R-07 — Where to call the awarding logic

**Decision**: `LoyaltyAwardUseCase.AwardForReceipt` is invoked from
`ReceiptConfirmUseCase.Execute` *after* the successful recompute step and
*before* the `EmitReceiptConfirmed` event. Awarding never aborts the
confirm: a failure in awarding logs and is swallowed (the receipt is
already `CONFIRMED` in DB; the price engine is already updated). The
caller can observe award results via the transaction logs.

**Rationale**: Keeps the existing successful confirm behavior intact and
treats points as the incentive layer, exactly like the spec model.

**Alternatives considered**:
- Async event-driven award (handler subscribed to
  `EmitReceiptConfirmed`): more decoupled but more complex; not required
  by the spec, violates Art. VI.1.

---

## R-08 — Balance: derived vs. stored

**Decision**: Keep `users.points` as the cached balance (already exists
and is already queried by the existing profile endpoint). Each award
performs `UPDATE users SET points = points + $delta` inside the same
transaction as the `loyalty_transactions` insert. The loyalty endpoint
optionally recomputes `SUM(loyalty_transactions.points)` as a self-check
in tests but returns the cached value to the user for performance. FR-011
requires the two to agree; a migration back-fills `users.points` from
`loyalty_transactions` is unnecessary because no points have been awarded
yet (the feature is new).

**Rationale**: O(1) balance read, satisfies SC-006 sub-second latency.

**Alternatives considered**:
- Drop `users.points` and always `SUM(...)`: slower for large histories.

---

## Summary

All open technical decisions are resolved without needing clarifications
from the user. The plan proceeds to Phase 1 design.