# Feature Specification: Receipt OCR Processing and Review

**Feature Branch**: `003-receipt-ocr-flow`

**Created**: 2026-06-24

**Status**: Draft

**Input**: User description: "Generate the technical plan for the receipt and OCR flow."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upload Receipt for Processing (Priority: P1)

An authenticated user uploads a receipt image and receives immediate confirmation that the receipt entered asynchronous processing.

**Why this priority**: Without reliable upload and queueing, the rest of the workflow cannot happen.

**Independent Test**: Upload one valid receipt image and verify immediate acceptance, receipt registration in pending state, and one processing task scheduled.

**Acceptance Scenarios**:

1. **Given** an authenticated user with a valid receipt image, **When** the user submits the upload, **Then** the system accepts the request immediately and returns a receipt identifier.
2. **Given** an accepted upload, **When** the receipt is queried right away, **Then** its lifecycle state is `PENDING`.
3. **Given** the same user uploads the same image again, **When** duplicate detection runs, **Then** the system returns the original receipt identifier and avoids duplicate downstream processing.

---

### User Story 2 - Review Parsed Receipt Data (Priority: P1)

After background extraction runs, the user can retrieve an editable summary containing merchant, date, total, and line items.

**Why this priority**: User review is required to convert OCR output into trusted purchase data.

**Independent Test**: Process a queued receipt and verify detail retrieval returns a reviewable summary in `NEEDS_REVIEW`, including partial results for low-quality scans.

**Acceptance Scenarios**:

1. **Given** a queued receipt, **When** background processing completes, **Then** the receipt transitions from `PENDING` to `NEEDS_REVIEW`.
2. **Given** a receipt in review state, **When** the user fetches receipt details, **Then** the response includes editable store, date, total, and line-item fields.
3. **Given** unreadable or incomplete extraction, **When** review data is returned, **Then** the receipt still becomes editable with available fields and no processing dead-end.

---

### User Story 3 - Confirm Corrected Receipt Data (Priority: P2)

The user confirms corrected receipt values so normalized product price observations can be stored and the receipt can be finalized.

**Why this priority**: Confirmation is the trust boundary for analytics and ranking data.

**Independent Test**: Confirm one reviewed receipt and verify corrected values are persisted, observations are created with required currency, and receipt transitions to final state.

**Acceptance Scenarios**:

1. **Given** a receipt in `NEEDS_REVIEW`, **When** the user submits corrected values, **Then** the receipt transitions to `CONFIRMED` and stores corrected data.
2. **Given** confirmed line items, **When** observations are saved, **Then** each observation includes a currency value and links to normalized product and store records.
3. **Given** missing currency on any confirmed line item, **When** confirmation is attempted, **Then** the system rejects confirmation and keeps the receipt in review state.

---

### Edge Cases

- How does the system respond when extraction returns no reliable line items?
- What happens when a store is not recognized in extracted text?
- How does the system prevent duplicate effects from repeated upload attempts of the same image by the same user?
- What happens when processing fails repeatedly for the same receipt?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST require authentication for receipt upload, receipt retrieval, and receipt confirmation.
- **FR-002**: System MUST accept receipt image uploads and respond asynchronously with an identifier for later retrieval.
- **FR-003**: System MUST store uploaded receipt images in object storage using non-guessable references.
- **FR-004**: System MUST register each accepted receipt in lifecycle state `PENDING`.
- **FR-005**: System MUST schedule OCR processing jobs in a queue and process them with concurrent workers.
- **FR-006**: System MUST keep text extraction behind a replaceable OCR provider contract so provider changes do not alter business behavior.
- **FR-007**: System MUST parse extracted receipt text into merchant, date, total, and line-item candidates through a decoupled, testable parser use case.
- **FR-008**: System MUST move receipts to `NEEDS_REVIEW` after a processing attempt, even when extraction quality is low, to allow manual completion.
- **FR-009**: System MUST expose a versioned API contract for upload, receipt detail retrieval, and confirmation under `/api/v1`.
- **FR-010**: System MUST enforce lifecycle transitions `PENDING -> NEEDS_REVIEW -> CONFIRMED | REJECTED`.
- **FR-011**: System MUST persist corrected store, date, total, and items during confirmation.
- **FR-012**: System MUST normalize confirmed items to canonical product identities before storing price observations.
- **FR-013**: System MUST persist price observations with mandatory currency and reject confirmation when any currency is missing.
- **FR-014**: System MUST support creation or association of previously unknown stores during review or confirmation.
- **FR-015**: System MUST provide schema support for receipts, receipt items, stores, products, and price observations with migration scripts.
- **FR-016**: System MUST include representative OCR-text fixtures from supermarket receipts to validate parser behavior in automated tests.

### Key Entities *(include if feature involves data)*

- **Receipt**: User-submitted purchase record with owner, image reference, lifecycle state, purchase date, total, and store association.
- **ReceiptItem**: Parsed or corrected line item containing raw text, quantity, unit price, currency, and canonical product mapping.
- **Store**: Merchant identity including name and optional branch/address used to group observations.
- **Product**: Canonical product definition used to normalize diverse receipt item names.
- **PriceObservation**: Confirmed product/store price capture that always includes currency and observation time.
- **OCRJob**: Asynchronous processing unit linking a receipt to extraction and parsing attempts.
- **RawOCRResult**: Structured extraction output returned by OCR providers before parser normalization.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: At least 95% of valid receipt uploads are acknowledged in under 2 seconds.
- **SC-002**: At least 95% of accepted receipts become reviewable (`NEEDS_REVIEW`) within 60 seconds under normal load.
- **SC-003**: 100% of confirmed price observations include currency.
- **SC-004**: 100% of same-user same-image duplicate uploads return the original receipt identifier without duplicate processing effects.
- **SC-005**: At least 90% of tested OCR-text fixture cases produce parser outputs that match expected merchant/date/item fields.

## Assumptions

- Existing authentication and user identity mechanisms are already available.
- Receipt uploads remain within existing payload size and file-type limits.
- Background processing capacity is sufficient to handle expected MVP traffic.
- Lifecycle and contract evolution beyond `/api/v1` is out of scope for this feature.
- Downstream analytics consumers are not implemented here; this feature only prepares validated observation data.
