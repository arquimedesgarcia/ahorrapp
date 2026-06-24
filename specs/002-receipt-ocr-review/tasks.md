# Tasks: Receipt OCR Review Flow

**Input**: Design documents from `/specs/002-receipt-ocr-review/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included. This feature requires strict behavior validation (async flow, state transitions, currency guarantees, provider swapability).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create feature scaffolding and shared contracts

- [ ] T001 Create receipt feature HTTP route scaffold in `internal/adapter/http/receipt_routes.go`
- [x] T002 Create receipt domain entities scaffold in `internal/domain/entities/receipt.go`
- [x] T003 [P] Create OCR parsing fixture directory with sample placeholder files in `internal/usecase/fixtures/ocr/`
- [x] T004 [P] Add feature migration index notes in `migrations/README.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core ports, repositories, queue and worker foundations required by all stories

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Define receipt-related domain ports (repository, queue, storage, OCR, event emitter) in `internal/domain/ports/receipt.go`
- [x] T006 Define receipt lifecycle/status constants and transition guards in `internal/domain/entities/receipt_status.go`
- [x] T007 [P] Implement PostgreSQL receipt repository base methods in `internal/adapter/postgres/receipt_repository.go`
- [x] T008 [P] Implement Redis queue adapter base methods for OCR jobs in `internal/adapter/redis/ocr_queue.go`
- [x] T009 [P] Implement object storage adapter methods for receipt image upload in `internal/adapter/storage/receipt_storage.go`
- [x] T010 [P] Implement OCR adapter client interface wrapper in `internal/adapter/ocr/paddle_client.go`
- [x] T011 Add receipt and OCR job migrations in `migrations/000002_receipts_ocr_jobs.up.sql`
- [x] T012 Add rollback for receipt and OCR job migrations in `migrations/000002_receipts_ocr_jobs.down.sql`

**Checkpoint**: Foundational components complete; user stories can proceed independently.

---

## Phase 3: User Story 1 - Upload receipt for async processing (Priority: P1) 🎯 MVP

**Goal**: Authenticated users can upload receipt images and immediately receive `202` with async processing enqueued.

**Independent Test**: `POST /api/v1/receipts` returns `202`, persists `PENDING` receipt with unguessable image URL, enqueues one OCR job, and duplicate same-user same-image upload returns existing receipt id without new job.

### Tests for User Story 1

- [x] T013 [P] [US1] Add unit tests for upload use case including duplicate-idempotency in `internal/usecase/receipt_upload_test.go`
- [ ] T014 [P] [US1] Add integration tests for `POST /api/v1/receipts` happy path and duplicate path in `internal/adapter/http/receipt_upload_handler_test.go`

### Implementation for User Story 1

- [x] T015 [US1] Implement receipt upload use case (store image, create `PENDING`, enqueue OCR job) in `internal/usecase/receipt_upload.go`
- [x] T016 [US1] Implement duplicate detection by `(user_id, image_hash)` and idempotent response in `internal/adapter/postgres/receipt_repository.go`
- [x] T017 [US1] Implement `POST /api/v1/receipts` handler contract mapping in `internal/adapter/http/receipt_upload_handler.go`
- [x] T018 [US1] Register authenticated upload route in `internal/adapter/http/receipt_routes.go`

**Checkpoint**: US1 independently functional and testable.

---

## Phase 4: User Story 2 - Receive editable parsed summary (Priority: P1)

**Goal**: Worker processes OCR jobs and users can fetch editable `NEEDS_REVIEW` summary.

**Independent Test**: After queued processing, `GET /api/v1/receipts/{id}` returns editable summary with store/date/total/items and receipt status `NEEDS_REVIEW` (including unreadable OCR fallback).

### Tests for User Story 2

- [ ] T019 [P] [US2] Add unit tests for OCR processing and parsing state transition rules in `internal/usecase/receipt_process_test.go`
- [ ] T020 [P] [US2] Add integration tests for `GET /api/v1/receipts/{id}` editable response in `internal/adapter/http/receipt_get_handler_test.go`

### Implementation for User Story 2

- [x] T021 [US2] Implement OCR worker loop and job processing orchestration in `internal/usecase/receipt_worker.go`
- [x] T022 [US2] Implement OCR text parser for merchant/date/total/items extraction in `internal/usecase/receipt_parser.go`
- [x] T023 [US2] Implement unreadable-image fallback to `NEEDS_REVIEW` with empty/partial items in `internal/usecase/receipt_process.go`
- [x] T024 [US2] Implement receipt detail use case for editable summary retrieval in `internal/usecase/receipt_get.go`
- [x] T025 [US2] Implement `GET /api/v1/receipts/{id}` handler in `internal/adapter/http/receipt_get_handler.go`
- [x] T026 [US2] Register receipt detail route in `internal/adapter/http/receipt_routes.go`

**Checkpoint**: US2 independently functional and testable.

---

## Phase 5: User Story 3 - Confirm corrected receipt and persist observations (Priority: P2)

**Goal**: Users confirm corrected receipt data, persist canonical mappings and currency-required observations, and transition to `CONFIRMED`.

**Independent Test**: `POST /api/v1/receipts/{id}/confirm` persists corrected values, creates canonical product links + `PriceObservation` records, enforces mandatory currency (validation failure when missing), emits downstream event calls, and sets status to `CONFIRMED` on success.

### Tests for User Story 3

- [x] T027 [P] [US3] Add unit tests for confirmation validation and observation creation in `internal/usecase/receipt_confirm_test.go`
- [ ] T028 [P] [US3] Add integration tests for `POST /api/v1/receipts/{id}/confirm` success and missing-currency failure in `internal/adapter/http/receipt_confirm_handler_test.go`

### Implementation for User Story 3

- [x] T029 [US3] Implement receipt confirmation use case with atomic persistence in `internal/usecase/receipt_confirm.go`
- [ ] T030 [US3] Implement canonical product normalization bridge in `internal/usecase/product_normalizer.go`
- [ ] T031 [US3] Implement `PriceObservation` persistence with mandatory currency constraint in `internal/adapter/postgres/price_observation_repository.go`
- [ ] T032 [US3] Implement unknown-merchant create/associate behavior in confirmation flow in `internal/usecase/store_resolution.go`
- [x] T033 [US3] Implement downstream event emission port calls (points + aggregate recompute) in `internal/usecase/receipt_confirm.go`
- [x] T034 [US3] Implement `POST /api/v1/receipts/{id}/confirm` handler in `internal/adapter/http/receipt_confirm_handler.go`
- [x] T035 [US3] Register confirm route in `internal/adapter/http/receipt_routes.go`

**Checkpoint**: US3 independently functional and testable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: End-to-end hardening, observability, and contract consistency

- [ ] T036 [P] Add structured logs/metrics around OCR job lifecycle and receipt transitions in `internal/usecase/receipt_worker.go`
- [ ] T037 [P] Add retry/backoff and dead-letter handling policy for OCR jobs in `internal/adapter/redis/ocr_queue.go`
- [ ] T038 [P] Update API contract examples and error payloads for final behavior in `specs/002-receipt-ocr-review/contracts/receipt-api-contract.md`
- [ ] T039 Update quickstart validation steps to match final implementation and edge cases in `specs/002-receipt-ocr-review/quickstart.md`
- [ ] T040 Run end-to-end validation checklist and record evidence in `specs/002-receipt-ocr-review/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: starts immediately.
- **Phase 2 (Foundational)**: depends on Phase 1; blocks all user stories.
- **Phase 3 (US1)**: depends on Phase 2.
- **Phase 4 (US2)**: depends on Phase 2 and uses US1 upload/queue output.
- **Phase 5 (US3)**: depends on Phase 2 and requires US2 review state flow.
- **Phase 6 (Polish)**: depends on US1+US2+US3 completion.

### User Story Dependencies

- **US1**: independent after Foundational; provides asynchronous ingestion baseline.
- **US2**: depends on US1-generated receipts/jobs to validate processing+review.
- **US3**: depends on US2 `NEEDS_REVIEW` outputs to confirm corrections.

### Within Each User Story

- Write tests first for that story.
- Implement use case logic before HTTP handler mapping.
- Register routes after handler implementation.
- Validate independent test criteria before moving to next story.

### Parallel Opportunities

- Foundational adapters `T007`–`T010` can run in parallel.
- US1 tests `T013` + `T014` can run in parallel.
- US2 tests `T019` + `T020` can run in parallel.
- US3 tests `T027` + `T028` can run in parallel.
- Polish tasks `T036`, `T037`, `T038` can run in parallel.

---

## Parallel Example: User Story 1

```bash
# Parallel test preparation
Task: "T013 [US1] internal/usecase/receipt_upload_test.go"
Task: "T014 [US1] internal/adapter/http/receipt_upload_handler_test.go"

# Then core implementation sequence
Task: "T015 [US1] internal/usecase/receipt_upload.go"
Task: "T017 [US1] internal/adapter/http/receipt_upload_handler.go"
```

## Parallel Example: User Story 3

```bash
# Parallel tests
Task: "T027 [US3] internal/usecase/receipt_confirm_test.go"
Task: "T028 [US3] internal/adapter/http/receipt_confirm_handler_test.go"

# Parallel supporting implementations
Task: "T030 [US3] internal/usecase/product_normalizer.go"
Task: "T032 [US3] internal/usecase/store_resolution.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (Setup)
2. Complete Phase 2 (Foundational)
3. Complete Phase 3 (US1)
4. Validate upload async behavior (`202`, `PENDING`, queued job)
5. Demo ingestion baseline

### Incremental Delivery

1. Deliver US1 (upload + queue)
2. Deliver US2 (OCR processing + editable summary)
3. Deliver US3 (confirmation + observations)
4. Apply polish/hardening

### Suggested MVP Scope

- **Strict MVP**: US1 only
- **Recommended practical MVP**: US1 + US2 (user-visible review workflow)

---

## Notes

- All tasks follow strict checklist format with IDs, optional `[P]`, required `[US#]` in story phases, and explicit file paths.
- Tests are included because spec requires testable acceptance for critical behavior and constraints.
