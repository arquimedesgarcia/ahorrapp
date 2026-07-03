# Feature Specification: Loyalty Points for Receipt Upload

**Feature Branch**: `006-loyalty-points`

**Created**: 2026-06-29

**Status**: Draft

**Input**: User description: "Feature: Loyalty system for uploading receipts. On confirming a valid receipt, the user earns points (configurable base amount, with possible bonuses for completing data or for being the first observation of a product/store). Each movement is recorded in LoyaltyTransaction (reason, points, date). Endpoint GET /api/v1/me/loyalty: point balance and history. Anti-abuse rules: the same receipt does not grant points twice; a configurable daily limit of point-granting receipts to mitigate fraud. Acceptance criteria: Confirming a new receipt grants points only once. The balance and history are queryable. Resubmitting the same receipt grants no additional points."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Earn points for a confirmed receipt (Priority: P1)

As a registered shopper, after I photograph a receipt, review the editable
summary, and confirm it as valid, I want to be awarded loyalty points so that
my contribution to the community price database is recognized and I am
incentivized to keep uploading.

**Why this priority**: This is the core of the gamification flywheel described
in the Global Design (value loop: uploads → data → recommendations → more
users). Without points awarded on confirmation, the loyalty feature does not
exist and the flywheel stalls.

**Independent Test**: A user can fully test it by uploading and confirming one
new receipt and verifying that their point balance increases by the expected
amount and that a single movement appears in their history. This delivers
standalone value (visible reward for contribution) even if bonuses or anti-
abuse rules are not yet exercised.

**Acceptance Scenarios**:

1. **Given** a registered user with zero points and a freshly uploaded receipt
   in `NEEDS_REVIEW`, **When** the user reviews and confirms the receipt
   (transition to `CONFIRMED`), **Then** the user's point balance increases by
   the configurable base amount and a `LoyaltyTransaction` is recorded with
   reason indicating base points for a confirmed receipt.
2. **Given** a confirmed receipt that has already granted points, **When** the
   same user (or any other user) submits the same receipt image or the same
   receipt record is re-confirmed, **Then** no additional points are granted
   and no new `LoyaltyTransaction` is created.
3. **Given** a user with a `PENDING` or `REJECTED` receipt, **When** the user
   attempts to confirm it, **Then** no points are granted because only
   `CONFIRMED` receipts are eligible.
4. **Given** the configurable base amount is changed by an administrator,
   **When** the next receipt is confirmed, **Then** the new base amount is
   applied (confirmations before the change are unaffected).

---

### User Story 2 - Earn bonus points for valuable contributions (Priority: P2)

As a contributor, when I confirm a receipt that completes missing data (e.g.,
adds a product not previously observed, or documents a store not previously
recorded), I want to earn bonus points on top of the base amount so that the
first observations — which are most valuable to the price database — are
incentivized.

**Why this priority**: Bonuses amplify data quality and coverage, but the
feature is functional with base points alone. Bonuses are the next layer of
value.

**Independent Test**: A user can confirm a receipt that introduces a brand-
new product or a brand-new store and verify the points awarded equal base
amount plus the corresponding bonus, with a separate `LoyaltyTransaction`
entry (or a single entry with a multi-part reason) describing each bonus.

**Acceptance Scenarios**:

1. **Given** no prior observation exists for product P at store S, **When**
   a user confirms a receipt containing P at S for the first time, **Then**
   the user earns the base points plus the configured "first product/store
   observation" bonus.
2. **Given** the user completes optional data fields on the editable summary
   (e.g., adds a missing purchase date or corrects a product unit), **When**
   the receipt is confirmed, **Then** the configured "data completion" bonus
   is awarded.
3. **Given** a receipt whose products and store were all previously observed,
   **When** it is confirmed, **Then** only the base points are awarded (no
   first-observation bonus).
4. **Given** the bonus configuration is modified, **When** a subsequent
   receipt is confirmed, **Then** only receipts confirmed after the change
   are affected.

---

### User Story 3 - Query point balance and history (Priority: P1)

As a user, I want to consult my current point balance and a chronological
history of every point movement (with the reason and date for each), so that
I can trust the system and understand how I earned each reward.

**Why this priority**: Transparency is required for trust in any points
system; the acceptance criteria explicitly require the balance and history to
be queryable. It is co-equal with the awarding of points itself for a usable
feature.

**Independent Test**: After confirming any number of receipts, the user can
retrieve an endpoint that returns their current balance and an ordered list
of every `LoyaltyTransaction` with reason, points, and date, and the values
match what was awarded.

**Acceptance Scenarios**:

1. **Given** a user with prior point movements, **When** the user requests
   their loyalty endpoint, **Then** the response includes the current point
   balance (sum of all movements) and an ordered history of movements, each
   with reason, points, and date.
2. **Given** a user with no movements, **When** the user requests their
   loyalty endpoint, **Then** the response shows a balance of zero and an
   empty history list (no error).
3. **Given** an authenticated user A and a separate user B with their own
   movements, **When** user A requests the endpoint, **Then** only user A's
   movements and balance are returned (no cross-user leakage).
4. **Given** a request without a valid authentication token, **When** the
   endpoint is called, **Then** the system rejects it as unauthenticated.

---

### Edge Cases

- **Duplicate receipt image**: The same image file is uploaded by the same or
  another user and confirmed. The duplicate is recognized as identical to an
  already-confirmed receipt and grants no points.
- **Daily cap reached**: A user has confirmed the configured maximum number of
  point-granting receipts in a single day (rolling or calendar day, per
  configuration). Subsequent confirmations on the same day still succeed (the
  receipt is valid and feeds the price engine) but earn **no** points, and the
  system records (or surfaces) that the daily cap was the reason.
- **Race condition on "first observation"**: Two users confirm receipts
  containing the same new product/store at nearly the same time. Only the
  first confirmation (by time of confirmation) earns the first-observation
  bonus; the second earns the base points only.
- **Receipt re-submission after rejection**: A previously rejected receipt
  image is uploaded again, corrected, and confirmed. It is treated as a new
  confirmation and earns points (the prior rejection does not antagonize a
  future valid confirmation of a corrected submission).
- **Configuration changes mid-day**: The daily cap or base amount changes
  during the day. The system applies the new values to confirmations that
  occur after the change; partially-awarded points earlier in the day are not
  retroactively recalculated.
- **Negative or zero-point edge**: A configuration value (base amount,
  bonus) is set to zero. The system still records a `LoyaltyTransaction`
  with zero points and the corresponding reason, so the history remains
  complete and explainable.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST award the configurable base points to the
  confirming user exactly once per confirmed (status `CONFIRMED`) receipt.
- **FR-002**: The system MUST record each point movement as a
  `LoyaltyTransaction` containing at minimum the user, the points
  (positive, zero, or negative if the system later permits deductions), a
  human-readable reason, and the date/time the movement occurred.
- **FR-003**: The system MUST expose an endpoint at `/api/v1/me/loyalty` that
  returns the authenticated user's current point balance and an ordered
  history of their point movements.
- **FR-004**: The system MUST prevent the same receipt (same receipt record,
  or a duplicate receipt image identified as identical to an already-confirmed
  receipt) from granting points more than once.
- **FR-005**: The system MUST enforce a configurable daily limit on the
  number of point-granting receipts per user. Confirmations beyond the limit
  must still mark the receipt as `CONFIRMED` and feed the price engine, but
  must not award points; the reason must reflect that the daily limit was
  reached.
- **FR-006**: The system MUST award a configurable "first observation"
  bonus to the user who confirms the first observation of a product at a
  given store (or first-ever observation of a store), recorded with a reason
  distinguishing it from base points.
- **FR-007**: The system MUST award a configurable "data completion"
  bonus when the user completes optional data fields on the editable summary
  before confirming the receipt, recorded with a reason distinguishing it.
- **FR-008**: The base amount and both bonus amounts and the daily limit
  MUST be configurable without code changes (e.g., via environment
  configuration or a configuration table) so administrators can tune the
  incentive economy.
- **FR-009**: The loyalty endpoint MUST return only the data of the
  authenticated requester; it MUST NOT expose another user's balance or
  movements.
- **FR-010**: The loyalty endpoint MUST reject requests that lack valid
  authentication.
- **FR-011**: The point balance MUST be derivable as the sum of all
  `LoyaltyTransaction` movements for the user, ensuring the balance and the
  history always agree.
- **FR-012**: The system MUST NOT award points for receipts in statuses other
  than `CONFIRMED` (i.e., `PENDING`, `NEEDS_REVIEW`, `REJECTED` are
  ineligible).
- **FR-013**: When a configuration value changes, the system MUST apply the
  new value to confirmations occurring after the change; it MUST NOT
  retroactively recalculate points awarded before the change.

### Key Entities *(include if feature involves data)*

- **User (existing entity, extended)**: Carries a `points` balance that
  reflects the sum of all `LoyaltyTransaction` movements for that user.
- **LoyaltyTransaction**: Records one point movement for a user. Key
  attributes: the user it belongs to, the number of points (can be negative
  or zero), the reason the movement occurred (base points, first-observation
  bonus, data-completion bonus, daily-limit-reached, future redemption, etc.),
  and the date/time the movement was created. Optionally references the
  receipt that triggered it.
- **Receipt (existing entity)**: Carries a flag or relation indicating
  whether loyalty points have already been granted for it, used to enforce
  the "award once" rule.
- **LoyaltyConfiguration**: A configurable set of values: base points per
  confirmed receipt, first-observation bonus, data-completion bonus, and
  daily point-granting-receipt cap per user. Values are changeable without
  code changes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of newly confirmed unique receipts grant points exactly
  once; resubmission of the same receipt grants zero additional points in
  every test case.
- **SC-002**: The daily-limit rule grants points on the configured number of
  receipts per user per day and grants zero points on every confirmation
  beyond that cap, with the reason recorded as daily-limit-reached.
- **SC-003**: A user can retrieve their complete point balance and the full
  ordered history of every movement, and the returned balance equals the sum
  of the returned movements in 100% of cases.
- **SC-004**: First-observation and data-completion bonuses are awarded
  correctly in the corresponding qualifying scenarios, and withheld in non-
  qualifying scenarios, in 100% of test cases.
- **SC-005**: A user with no prior activity sees a zero balance and an empty
  history (rather than an error) on the loyalty endpoint on the first
  attempt.
- **SC-006**: The point-balance and history query returns within a perceptibly
  instantaneous time for a typical user (under 1 second of waiting) so the
  rewards screen feels responsive.
- **SC-007**: Anonymous (unauthenticated) access to the loyalty endpoint is
  rejected in 100% of attempts, and no user can see another user's balance or
  movements.

## Assumptions

- The receipt-lifecycle states (`PENDING`, `NEEDS_REVIEW`, `CONFIRMED`,
  `REJECTED`) and the act of confirming a receipt through the editable
  summary already exist or are delivered by the receipt-upload feature
  (Epics E3–E6); this feature layers point awarding on top of the
  `CONFIRMED` transition.
- Authentication is JWT-based (per Constitution Article VII) and the
  authenticated user identity is available to the loyalty endpoint via the
  existing auth mechanism; no new auth scheme is introduced.
- The data model already defines `User`, `Receipt`, `Product`, `Store`, and
  `LoyaltyTransaction` entities as in the Global Design; this feature uses
  and possibly extends `LoyaltyTransaction` and adds configuration.
- "First observation" is determined at confirmation time by whether a
  `PriceObservation` for the same (product, store) pair — or, for the store
  bonus, any prior `Store`/`PriceObservation` for that store — already
  existed before the current confirmation.
- The daily limit is applied per user, per calendar day in the user's
  timezone (or UTC by default) — the exact day boundary is an implementation
  detail decided in the plan; both are reasonable defaults.
- Configuration values are stored in a way the project already uses for
  tunable settings (environment variables or a configuration table); the
  chosen mechanism is a plan-level decision.
- Duplicate receipt detection relies on a mechanism provided or configured
  at the plan level (e.g., an image hash on the receipt); this feature only
  depends on the ability to identify a duplicate as "the same receipt" and
  block double-granting.
- Only point awards from confirmed receipts, the two documented bonuses, and
  the daily-limit rule are in scope for this feature. Redemption of points,
  spending, leaderboards, and reward tiers are NOT in scope for this feature
  (they belong to the broader Gamification epic E9 and future features).