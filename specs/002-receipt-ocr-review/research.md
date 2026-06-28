# Research: Receipt OCR Review Flow

## Decision 1: OCR Provider Contract Shape

**Decision**: Use a domain port that accepts image reference and returns normalized OCR text blocks plus metadata.

**Rationale**: Keeps use-case logic independent from PaddleOCR-specific payload formats and allows provider swaps.

**Alternatives considered**:
- Returning provider-native JSON directly: rejected (leaks provider schema into use cases).
- Parsing directly inside adapter with no intermediate model: rejected (harder to test parsing logic independently).

## Decision 2: Async Processing Model

**Decision**: Enqueue one OCR job per accepted upload in Redis; worker processes jobs and updates receipt state.

**Rationale**: Immediate API response with deferred heavy OCR work; aligns with queue-based architecture in global design.

**Alternatives considered**:
- Synchronous OCR in upload request: rejected (latency spikes, poor UX, fragile mobile network behavior).
- External queue service: rejected (violates local-first MVP cost goals).

## Decision 3: Duplicate Handling Strategy

**Decision**: Same-user same-image duplicates are idempotent: return existing receipt ID and do not enqueue new job.

**Rationale**: Prevents duplicated processing costs and downstream side effects while preserving deterministic API behavior.

**Alternatives considered**:
- Create duplicate records with dedupe downstream: rejected (complexity and noisy analytics).
- Hard reject with 409: rejected (worse UX than idempotent recovery path).

## Decision 4: Unreadable OCR Handling

**Decision**: Unreadable OCR output still transitions receipt to `NEEDS_REVIEW` with editable empty/partial items.

**Rationale**: Preserves manual correction path and avoids dead-end failures.

**Alternatives considered**:
- Mark as failed terminal state: rejected (blocks user correction workflow).

## Decision 5: Confirmation Validation for Currency

**Decision**: Currency is mandatory for every persisted `PriceObservation`; confirmation fails validation for any item missing currency.

**Rationale**: Enforces constitution Article V (no mixed/implicit currency).

**Alternatives considered**:
- Defaulting missing currency from store heuristic: rejected (can silently corrupt price quality).

## Decision 6: Merchant Resolution

**Decision**: If merchant cannot be resolved, allow creating/associating a new Store during review/confirmation.

**Rationale**: Matches expected behavior and keeps user flow unblocked.

**Alternatives considered**:
- Require pre-existing merchant only: rejected (high friction and frequent OCR mismatch failures).

## Decision 7: Event Emission Scope

**Decision**: On confirmation, emit domain events/calls for loyalty and aggregate recomputation, but do not implement handlers in this feature.

**Rationale**: Keeps scope bounded while preserving integration contract with later epics.

**Alternatives considered**:
- Implement full loyalty and aggregate updates now: rejected (out of current feature scope).
