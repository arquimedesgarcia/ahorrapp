# Research: Receipt OCR Processing and Review

## Decision 1: Redis Queue + Concurrent Worker

- **Decision**: Use Redis list-backed queue for OCR jobs with concurrent Go workers.
- **Why**:
  - Simple operationally in local-first Docker stack.
  - Matches current infrastructure and avoids extra broker complexity.
  - Supports practical retry and dead-letter patterns.
- **Alternatives considered**:
  - PostgreSQL polling queue: rejected due to higher contention and latency.
  - Dedicated brokers: rejected for MVP complexity and cost.

## Decision 2: OCR Provider Port Contract

- **Decision**: Domain port signature:

```go
type OCRProvider interface {
    Extract(ctx context.Context, imageRef string) (RawOCRResult, error)
}
```

- **Why**:
  - Keeps extraction provider swappable per constitution Article I.
  - Encapsulates OCR output in a stable domain structure (`RawOCRResult`).

## Decision 3: PaddleOCR Adapter via Separate Python Service

- **Decision**: Implement `PaddleOCRProvider` in Go adapter layer, sending HTTP POST to a separate FastAPI + PaddleOCR service.
- **Why**:
  - Separates heavyweight OCR runtime from core API image.
  - Allows independent scaling and updates without touching core domain logic.
- **Service minimum endpoints**:
  - `POST /extract` with `image_ref`
  - `GET /health`

## Decision 4: S3-Compatible Storage Adapter

- **Decision**: One `StorageProvider` adapter for S3-compatible APIs; target differs by configuration only.
- **Development target**: MinIO container in Docker Compose.
- **Production target**: Hetzner Object Storage.
- **Why**:
  - Satisfies constitution requirement for replaceable details and local-first.
  - Avoids environment-specific code branches.

## Decision 5: Parser as Decoupled Domain Use Case

- **Decision**: Keep receipt parser in use case layer with fixture-based tests.
- **Fixture strategy**:
  - Supermarket OCR text examples (readable, partial, unreadable).
  - Expected parsed fields for store/date/total/items.
- **Why**:
  - Deterministic tests without external OCR dependency.
  - Supports incremental parser improvements safely.

## Decision 6: State Machine and Validation

- **Decision**: Enforce transitions `PENDING -> NEEDS_REVIEW -> CONFIRMED | REJECTED`.
- **Why**:
  - Prevents invalid lifecycle jumps.
  - Makes queue/processing and confirmation behavior auditable.
