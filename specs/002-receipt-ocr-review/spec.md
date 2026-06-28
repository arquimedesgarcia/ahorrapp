# Feature Specification: Receipt OCR Review Flow

**Feature Branch**: `002-receipt-ocr-review`

**Created**: 2026-06-21

**Status**: Draft

**Input**: User description: "Feature: receipt upload, OCR processing, parsing, and editable review. Authenticated users upload receipt images, processing is asynchronous, OCR is swappable behind OCRProvider, parsed data moves receipts to review, users edit and confirm, and confirmations create normalized price observations with mandatory currency."

## Clarifications

### Session 2026-06-21

- Q: How should same-user same-image duplicate uploads be handled? → A: Treat as idempotent: return existing receipt ID with duplicate flag; create no new receipt/job.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upload receipt for async processing (Priority: P1)

An authenticated user submits a receipt image and gets an immediate response that their upload was accepted for processing, without waiting for OCR completion.

**Why this priority**: This is the entry point for all receipt intelligence. If upload and queueing fail, no OCR or review workflow can happen.

**Independent Test**: Submit an authenticated receipt upload request and verify the API returns `202` with a receipt ID, stores the image using an unguessable URL, creates the receipt in `PENDING`, and queues one processing job.

**Acceptance Scenarios**:

1. **Given** an authenticated user and a valid image file, **When** the user uploads a receipt,
   **Then** the system returns `202 Accepted` with a receipt identifier and does not block while OCR runs.
2. **Given** an accepted upload, **When** the receipt record is retrieved immediately,
   **Then** its status is `PENDING` and it is linked to the uploader.
3. **Given** an accepted upload, **When** the processing queue is inspected,
   **Then** exactly one OCR-processing job exists for that receipt.
4. **Given** the same user uploads the same image again, **When** the duplicate is detected,
   **Then** the system returns the existing receipt ID with duplicate indication and does not create a new receipt or queue job.

---

### User Story 2 - Receive editable parsed summary (Priority: P1)

After background OCR processing, the user can view an editable text summary containing detected store,
date, total, and line items.

**Why this priority**: Parsed output must be human-correctable before it can become trusted price data.

**Independent Test**: Upload a receipt, wait for processing, call receipt detail endpoint, and verify the receipt reaches `NEEDS_REVIEW` with editable fields for store, date, total, and items.

**Acceptance Scenarios**:

1. **Given** a queued receipt and a readable image, **When** the background worker completes OCR and parsing,
   **Then** the receipt transitions from `PENDING` to `NEEDS_REVIEW`.
2. **Given** a receipt in `NEEDS_REVIEW`, **When** the app fetches receipt details,
   **Then** the response includes editable store, purchase date, total, and line-item fields.
3. **Given** OCR provider implementation A is replaced by implementation B,
   **When** processing the same receipt flow,
   **Then** use-case behavior remains unchanged and no use-case code modification is required.

---

### User Story 3 - Confirm corrected receipt and persist observations (Priority: P2)

The user edits parsed content and confirms it, producing canonical product mappings and price observations with mandatory currency.

**Why this priority**: Confirmation is the trust boundary where user-validated data becomes analytics input.

**Independent Test**: Confirm a reviewed receipt with corrected payload and verify persisted corrections, canonical product mapping, `PriceObservation` creation per line item, and receipt status transition to `CONFIRMED`.

**Acceptance Scenarios**:

1. **Given** a receipt in `NEEDS_REVIEW`, **When** the user submits corrected receipt content,
   **Then** the system persists the corrected store/date/total/items and marks the receipt `CONFIRMED`.
2. **Given** corrected line items, **When** confirmation completes,
   **Then** each line item is linked to a canonical product and produces one `PriceObservation`.
3. **Given** created `PriceObservation` records, **When** they are stored,
   **Then** each record includes a non-empty currency and no observation is stored without currency.

---

### Edge Cases

- **Unreadable image**: If OCR cannot read usable content, receipt still transitions to `NEEDS_REVIEW` with empty or partial line items so user can complete data manually.
- **Unrecognized store/merchant**: If parsed merchant is unknown, system allows creating a new store entry during review/confirmation.
- **Duplicate receipt**: If the same user uploads the same image again, system treats the request as idempotent by returning the existing receipt ID (with duplicate indication), creating no new receipt, and enqueuing no new OCR job.
- **Partial OCR extraction**: If OCR extracts date/total but misses line items, user still receives editable summary with available fields and can complete missing data manually.
- **Currency ambiguity in line items**: If OCR text does not clearly include currency, summary remains editable but confirmation cannot finalize any item without explicit currency value.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST require authentication for upload, receipt detail, and receipt confirmation actions.
- **FR-002**: System MUST accept receipt image uploads and return `202 Accepted` with a receipt identifier for asynchronous processing.
- **FR-003**: On accepted upload, system MUST store the image using an unguessable retrieval reference.
- **FR-004**: On accepted upload, system MUST create a receipt record in `PENDING` status.
- **FR-005**: On accepted upload, system MUST enqueue exactly one OCR-processing job for that receipt.
- **FR-006**: A background worker MUST process queued OCR jobs independently of request/response lifecycle.
- **FR-007**: OCR processing MUST use an `OCRProvider` abstraction so provider implementation can be replaced without changing use-case behavior.
- **FR-008**: OCR parsing MUST extract merchant/store, purchase date, total, and line-item candidates (raw text, quantity, unit price, currency when available).
- **FR-009**: After OCR parsing finishes, system MUST move receipt status to `NEEDS_REVIEW` and expose editable summary fields.
- **FR-010**: System MUST provide receipt detail retrieval with editable summary content for `NEEDS_REVIEW` receipts.
- **FR-011**: System MUST accept corrected summary content and persist user-confirmed values on confirmation.
- **FR-012**: Confirmation MUST normalize each confirmed line item to a canonical product identity.
- **FR-013**: Confirmation MUST create one `PriceObservation` per confirmed line item with product, store, price, and mandatory currency.
- **FR-014**: No `PriceObservation` MUST be persisted without currency.
- **FR-015**: Successful confirmation MUST transition receipt status to `CONFIRMED`.
- **FR-016**: Confirmation MUST emit downstream domain events/calls for points and aggregate recomputation without implementing those epics in this feature.
- **FR-017**: For unreadable receipts, system MUST still produce a reviewable `NEEDS_REVIEW` receipt allowing manual completion.
- **FR-018**: System MUST support unknown-merchant handling by allowing creation/association of a new store during review/confirmation.
- **FR-019**: System MUST detect same-user same-image duplicates and handle them idempotently by returning the existing receipt ID (with duplicate indication) while creating no new receipt record and no new OCR job.

### Key Entities *(include if feature involves data)*

- **Receipt**: User-submitted purchase document with lifecycle states (`PENDING`, `NEEDS_REVIEW`, `CONFIRMED`, `REJECTED`), image reference, purchase date, total, and store association.
- **ReceiptItem**: Parsed or corrected line item with raw text, quantity, unit price, currency, and canonical product link after confirmation.
- **Store**: Merchant identity parsed from receipt (name, optional branch/address) used to group price observations by merchant.
- **Product**: Canonical product identity used to normalize varied raw item names.
- **PriceObservation**: Confirmed product/store price record containing mandatory currency and observation timestamp.
- **OCRJob**: Asynchronous processing unit linking a receipt to OCR extraction/parsing workflow and processing outcome metadata.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 95% of valid receipt uploads return acceptance response in under 2 seconds.
- **SC-002**: 95% of accepted uploads reach either `NEEDS_REVIEW` or an explicit processing-failure reviewable state within 60 seconds.
- **SC-003**: 100% of confirmed receipts persist user-edited values exactly as submitted for editable fields.
- **SC-004**: 100% of persisted `PriceObservation` records include currency.
- **SC-005**: Replacing OCR provider implementation requires zero changes to receipt-processing use-case code.
- **SC-006**: Duplicate-upload detection prevents duplicate downstream effects in 100% of same-user same-image duplicate attempts.

## Assumptions

- Backend skeleton and authentication flows are already implemented and available.
- Receipt image size/type limits and anti-malware scanning are handled by existing upload guardrails.
- OCR provider returns text payloads in a format mappable to store/date/total/items with parser fallback behavior.
- Duplicate detection uses a deterministic image fingerprint scoped by user identity.
- Manual review UI can supply full corrected payload including currency when OCR output is incomplete.
- Product normalization reuses canonical product records or creates candidates according to existing normalization policy.
- Event emission for points and aggregate recomputation is limited to signal dispatch in this feature; consumers are implemented in later epics.
