# Implementation Plan: Receipt OCR Review Flow

**Branch**: `002-receipt-ocr-review` | **Date**: 2026-06-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-receipt-ocr-review/spec.md`

## Summary

Implement end-to-end receipt ingestion and review workflow: authenticated image upload,
asynchronous OCR processing, merchant/date/item parsing, editable review retrieval, and
user confirmation that persists canonical product mappings and currency-required
`PriceObservation` records. The design keeps OCR and storage behind replaceable ports,
uses a queue-backed worker for async processing, and enforces receipt state transitions
(`PENDING -> NEEDS_REVIEW -> CONFIRMED`).

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**:
- chi (`net/http`-compatible routing)
- pgx (PostgreSQL access)
- go-redis (queue + cache client)
- minio-go (S3-compatible storage adapter)
- golang-migrate (versioned migrations)
- Optional parsing helpers for OCR text normalization (stdlib-first)

**Storage**:
- PostgreSQL 16 + PostGIS (core entities + relations)
- Redis (async OCR job queue)
- MinIO S3-compatible bucket (receipt images)

**Testing**:
- Unit tests for use cases (`upload`, `process`, `confirm`)
- Integration tests for critical endpoints (`POST /receipts`, `GET /receipts/{id}`, `POST /receipts/{id}/confirm`)

**Target Platform**: Dockerized Linux services for local-first development

**Project Type**: Backend web service + background worker

**Performance Goals**:
- Upload request returns `202` within 2s at p95
- OCR processing reaches `NEEDS_REVIEW` within 60s at p95 under nominal load

**Constraints**:
- No business logic depends on OCR provider implementation details
- Every `PriceObservation` must include currency
- Unreadable OCR results must still produce editable review state
- Duplicate same-user same-image uploads are idempotent (return existing receipt, no new job)

**Scale/Scope**:
- MVP flow for receipt upload/review only
- No loyalty-points implementation here (emit event/call only)
- No aggregate recomputation implementation here (emit event/call only)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Article I — Clean Architecture in Go backend**
  - Domain defines ports: `OCRProvider`, `StorageProvider`, queue/repository ports.
  - Use cases depend only on domain ports and entities.
  - Adapters implement ports (`postgres`, `redis`, `storage`, `ocr`, `http`).
  - **Status**: PASS.

- **Article IV — Explicit, versioned contracts**
  - All endpoints under `/api/v1/...`.
  - Request/response and error contracts documented in `contracts/receipt-api-contract.md`.
  - **Status**: PASS.

- **Article VI — Simplicity, cost, local-first**
  - Uses existing local Docker stack (API, Postgres, Redis, MinIO, OCR service).
  - Self-hosted PaddleOCR for MVP (no paid OCR dependency).
  - **Status**: PASS.

- **Article V — Data, currency, normalization**
  - Store extraction and canonical product normalization included.
  - Currency mandatory for all persisted `PriceObservation` records.
  - **Status**: PASS.

**Gate Result**: PASS — no unjustified violations.

## Project Structure

### Documentation (this feature)

```text
specs/002-receipt-ocr-review/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── receipt-api-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── api/

internal/
├── domain/
│   ├── entities/
│   └── ports/
├── usecase/
├── adapter/
│   ├── http/
│   ├── postgres/
│   ├── redis/
│   ├── storage/
│   └── ocr/
└── config/

migrations/
```

**Structure Decision**:
- Keep single service + internal worker process model for MVP.
- Keep parser as use-case-support component, testable via OCR fixtures.
- Keep event emission behind domain port (no direct coupling to future loyalty/aggregate modules).

## Phase 0: Outline & Research

Research focuses:
1. OCR provider adapter contract and timeout/retry strategy
2. Redis queue pattern for exactly-once-enough processing semantics
3. Receipt parsing strategy for noisy OCR text (merchant/date/items)
4. Duplicate detection using content hash + user scope
5. Confirmation validation strategy for mandatory currency

Output: `research.md` with decisions and alternatives.

## Phase 1: Design & Contracts

1. Define data entities and lifecycle transitions (`data-model.md`)
2. Define external API contracts for receipt flow (`contracts/receipt-api-contract.md`)
3. Define runnable validation flow (`quickstart.md`)
4. Update agent context pointer in `AGENTS.md`

Post-design Constitution re-check: PASS.

## File Creation Order

1. `research.md`
2. `data-model.md`
3. `contracts/receipt-api-contract.md`
4. `quickstart.md`
5. `AGENTS.md` plan pointer update

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
