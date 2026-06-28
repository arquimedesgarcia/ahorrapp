# Tasks: Receipt OCR Processing and Review

**Input**: Design documents from `/specs/003-receipt-ocr-flow/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included. This feature requires parser, use-case, contract, and integration verification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create feature scaffolding and baseline files

- [x] T001 Create receipt route registration scaffold in `internal/adapter/http/receipt_routes.go`
- [x] T002 Create OCR supermarket fixture files in `internal/usecase/fixtures/ocr/`
- [x] T003 [P] Add feature docs stubs in `specs/003-receipt-ocr-flow/contracts/`
- [x] T004 [P] Add migration index note for receipt OCR feature in `migrations/README.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core ports/adapters/schema needed by all user stories

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Define `RawOCRResult` and OCR provider port in `internal/domain/ports/ocr.go`
- [x] T006 Define receipt state transitions and guards in `internal/domain/entities/receipt_status.go`
- [x] T007 [P] Implement Redis queue adapter with retry/backoff and DLQ in `internal/adapter/redis/ocr_queue.go`
- [x] T008 [P] Implement S3-compatible storage adapter configuration in `internal/adapter/storage/client.go`
- [x] T009 [P] Implement `PaddleOCRProvider` adapter against OCR service contract in `internal/adapter/ocr/paddle_provider.go`
- [x] T010 Add DB migration for `receipts`, `receipt_items`, `stores`, `products`, `price_observations`, `ocr_jobs` in `migrations/000002_receipts_ocr_jobs.up.sql`
- [x] T011 Add rollback migration for receipt OCR schema in `migrations/000002_receipts_ocr_jobs.down.sql`
- [x] T012 Add MinIO service and S3 env var wiring in `docker-compose.yml` and `.env.example`
- [x] T013 Create minimal OCR Python microservice (`FastAPI` + `PaddleOCR`) in `ocr-service/`

**Checkpoint**: Foundation complete; stories can proceed.

---

## Phase 3: User Story 1 - Upload Receipt for Processing (Priority: P1) 🎯 MVP

**Goal**: Users upload receipt images and receive async acceptance with duplicate idempotency.

**Independent Test**: `POST /api/v1/receipts` returns `202`, persists `PENDING`, enqueues job, and duplicate upload returns existing `receipt_id`.

### Tests for User Story 1

- [x] T014 [P] [US1] Add upload use-case tests for create + duplicate behavior in `internal/usecase/receipt_upload_test.go`
- [x] T015 [P] [US1] Add HTTP integration tests for upload happy/duplicate paths in `internal/adapter/http/receipt_upload_handler_test.go`

### Implementation for User Story 1

- [x] T016 [US1] Implement upload use case (`store image`, `create PENDING`, `enqueue`) in `internal/usecase/receipt_upload.go`
- [x] T017 [US1] Implement duplicate detection `(user_id, image_hash)` in `internal/adapter/postgres/receipt_repository.go`
- [x] T018 [US1] Implement `POST /api/v1/receipts` handler in `internal/adapter/http/receipt_upload_handler.go`
- [x] T019 [US1] Register upload route under `/api/v1` in `internal/adapter/http/receipt_routes.go`

**Checkpoint**: US1 independently functional.

---

## Phase 4: User Story 2 - Review Parsed Receipt Data (Priority: P1)

**Goal**: Background OCR + parser produce editable receipt summary in `NEEDS_REVIEW`.

**Independent Test**: Worker processes job and `GET /api/v1/receipts/{id}` returns editable fields; unreadable OCR still returns reviewable output.

### Tests for User Story 2

- [x] T020 [P] [US2] Add parser unit tests using supermarket fixtures in `internal/usecase/receipt_parser_test.go`
- [x] T021 [P] [US2] Add OCR processing transition tests in `internal/usecase/receipt_process_test.go`
- [x] T022 [P] [US2] Add `GET /api/v1/receipts/{id}` integration test in `internal/adapter/http/receipt_get_handler_test.go`

### Implementation for User Story 2

- [x] T023 [US2] Implement parser use case for store/date/total/items in `internal/usecase/receipt_parser.go`
- [x] T024 [US2] Implement worker orchestration with concurrent consumption in `internal/usecase/receipt_worker.go`
- [x] T025 [US2] Implement OCR processing use case and fallback behavior in `internal/usecase/receipt_process.go`
- [x] T026 [US2] Implement receipt detail retrieval use case in `internal/usecase/receipt_get.go`
- [x] T027 [US2] Implement `GET /api/v1/receipts/{id}` handler in `internal/adapter/http/receipt_get_handler.go`
- [x] T028 [US2] Register receipt detail route in `internal/adapter/http/receipt_routes.go`

**Checkpoint**: US2 independently functional.

---

## Phase 5: User Story 3 - Confirm Corrected Receipt Data (Priority: P2)

**Goal**: Confirm corrected values, normalize products, persist observations with mandatory currency, and finalize lifecycle.

**Independent Test**: `POST /api/v1/receipts/{id}/confirm` persists corrected values, rejects missing currency, and transitions to `CONFIRMED`.

### Tests for User Story 3

- [x] T029 [P] [US3] Add confirmation unit tests for currency validation and state transition in `internal/usecase/receipt_confirm_test.go`
- [x] T030 [P] [US3] Add confirmation HTTP integration tests for success/failure in `internal/adapter/http/receipt_confirm_handler_test.go`

### Implementation for User Story 3

- [x] T031 [US3] Implement confirmation use case orchestration in `internal/usecase/receipt_confirm.go`
- [x] T032 [US3] Implement product normalization bridge in `internal/usecase/product_normalizer.go`
- [x] T033 [US3] Implement store resolution bridge for unknown merchants in `internal/usecase/store_resolution.go`
- [x] T034 [US3] Implement price observation persistence adapter in `internal/adapter/postgres/price_observation_repository.go`
- [x] T035 [US3] Implement `POST /api/v1/receipts/{id}/confirm` handler in `internal/adapter/http/receipt_confirm_handler.go`
- [x] T036 [US3] Register confirmation route in `internal/adapter/http/receipt_routes.go`

**Checkpoint**: US3 independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final contracts, docs, verification, and operational hardening

- [x] T037 [P] Add structured worker logs/metrics for queue lifecycle in `internal/usecase/receipt_worker.go`
- [x] T038 [P] Finalize public contract examples and errors in `specs/003-receipt-ocr-flow/contracts/receipt-api-contract.md`
- [x] T039 [P] Finalize OCR service contract examples in `specs/003-receipt-ocr-flow/contracts/ocr-service-contract.md`
- [x] T040 Update quickstart with MinIO bucket/env setup and end-to-end flow in `specs/003-receipt-ocr-flow/quickstart.md`
- [x] T041 Run full validation (`go test ./...`, docker compose smoke flow) and record evidence in `specs/003-receipt-ocr-flow/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: starts immediately.
- **Phase 2**: depends on Phase 1 and blocks all stories.
- **US1 (Phase 3)**: depends on Phase 2.
- **US2 (Phase 4)**: depends on Phase 2 and receipt jobs from US1.
- **US3 (Phase 5)**: depends on Phase 2 and review outputs from US2.
- **Phase 6**: depends on completion of required stories.

### User Story Dependencies

- **US1**: core ingestion baseline.
- **US2**: depends on ingestion + processing pipeline.
- **US3**: depends on review state and parsed/corrected data flow.

### Parallel Opportunities

- Foundational adapters `T007`-`T009` can run in parallel.
- US1 tests `T014` + `T015` can run in parallel.
- US2 tests `T020`-`T022` can run in parallel.
- US3 tests `T029` + `T030` can run in parallel.
- Contract/doc polish `T038` + `T039` + `T040` can run in parallel.

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Deliver US1 (`upload -> PENDING -> job queued`).
3. Validate duplicate idempotency and async acceptance.

### Incremental Delivery

1. Add US2 for reviewable parsed output.
2. Add US3 for confirmation and observation persistence.
3. Finalize contracts/docs and run full validation.
