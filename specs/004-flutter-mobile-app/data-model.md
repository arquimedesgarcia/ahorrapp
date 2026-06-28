# Data Model: Flutter Mobile Application MVP

Flutter-side data models that mirror the backend API contracts exactly. These models are used in the data layer (API clients) and transformed into domain entities for the presentation layer.

## Auth Models

### LoginRequest

- **Purpose**: POST /api/v1/auth/login request body.
- **Fields**:
  - `email`: String
  - `password`: String

### RegisterRequest

- **Purpose**: POST /api/v1/auth/register request body.
- **Fields**:
  - `email`: String
  - `password`: String
  - `display_name`: String

### AuthResponse

- **Purpose**: POST /api/v1/auth/login and /api/v1/auth/register response.
- **Fields**:
  - `token`: String (JWT)
  - `user`: UserProfile

### UserProfile

- **Purpose**: Current authenticated user identity.
- **Fields**:
  - `id`: String (UUID)
  - `email`: String
  - `display_name`: String

## Receipt Models

### ReceiptUploadResponse

- **Purpose**: POST /api/v1/receipts response.
- **Fields**:
  - `receipt_id`: String (UUID)
  - `status`: String ("PENDING")
  - `duplicate`: bool

### ReceiptDetail

- **Purpose**: GET /api/v1/receipts/{id} response (NEEDS_REVIEW state).
- **Fields**:
  - `receipt_id`: String (UUID)
  - `status`: String ("NEEDS_REVIEW")
  - `store`: StoreInfo
  - `purchase_date`: String (ISO date, nullable)
  - `total`: double (nullable)
  - `items`: List<ReceiptItem>

### StoreInfo

- **Purpose**: Store details within receipt context.
- **Fields**:
  - `name`: String (nullable)
  - `branch`: String (nullable)
  - `address`: String (nullable)

### ReceiptItem

- **Purpose**: Line item within receipt detail.
- **Fields**:
  - `raw_text`: String
  - `quantity`: int (nullable)
  - `unit_price`: double (nullable)
  - `currency`: String (nullable, "USD" or "Bs.")

### ConfirmReceiptRequest

- **Purpose**: POST /api/v1/receipts/{id}/confirm request body.
- **Fields**:
  - `store`: StoreInfo
  - `purchase_date`: String (ISO date)
  - `total`: double
  - `items`: List<ReceiptItem>

### ConfirmReceiptResponse

- **Purpose**: POST /api/v1/receipts/{id}/confirm success (204 No Content or 200 with points).
- **Fields**:
  - `points_earned`: int

### ReceiptStatus (Domain Entity)

- **Purpose**: Lifecycle state enum for UI logic.
- **Values**: `pending`, `needsReview`, `confirmed`, `rejected`
- **Mapping**: Mirrors backend state machine `PENDING -> NEEDS_REVIEW -> CONFIRMED | REJECTED`.

## Ranking Models

### ProductSearchRequest

- **Purpose**: GET /api/v1/ranking/products/search query parameters.
- **Fields**:
  - `q`: String (search query)

### ProductSearchResponse

- **Purpose**: GET /api/v1/ranking/products/search response.
- **Fields**:
  - `results`: List<ProductSearchResult>

### ProductSearchResult

- **Purpose**: One product match with store rankings.
- **Fields**:
  - `product_id`: String (UUID)
  - `product_name`: String
  - `unit`: String (nullable, "kg", "lt", "unit")
  - `stores`: List<StorePriceEntry> (ranked cheapest first)

### StorePriceEntry

- **Purpose**: A store's price for a product.
- **Fields**:
  - `store_id`: String (UUID)
  - `store_name`: String
  - `branch`: String (nullable)
  - `average_price`: double
  - `currency`: String ("USD" or "Bs.")
  - `sample_count`: int

## Profile Models

### PointsResponse

- **Purpose**: GET /api/v1/users/me/points response.
- **Fields**:
  - `total_points`: int
  - `recent_transactions`: List<PointsTransaction> (optional, for activity display)

### PointsTransaction

- **Purpose**: Points earned/lost event.
- **Fields**:
  - `id`: String (UUID)
  - `points`: int
  - `reason`: String
  - `created_at`: String (ISO datetime)

## UI State Models (Domain Entities)

### UploadState

- **Purpose**: Tracks receipt upload progress for UI.
- **States**:
  - `idle` — Camera ready, no upload in progress.
  - `capturing` — Camera viewfinder active.
  - `uploading(progress: double)` — Multipart upload in progress (0.0 to 1.0).
  - `retrying(attempt: int)` — Dio retry interceptor active.
  - `success(receiptId: String)` — Upload accepted, receipt ID available.
  - `duplicate(receiptId: String)` — Duplicate detected.
  - `error(message: String)` — Upload failed after retries.

### ReviewState

- **Purpose**: Tracks receipt review and confirmation flow.
- **States**:
  - `loading` — Fetching receipt details.
  - `ready(detail: ReceiptDetail)` — Editable summary displayed.
  - `confirming` — Submitting confirmation.
  - `confirmed(points: int)` — Confirmation succeeded, points displayed.
  - `error(message: String)` — Confirmation failed.

### SearchState

- **Purpose**: Tracks product search flow.
- **States**:
  - `idle` — No search performed yet.
  - `loading` — Search in progress.
  - `results(List<ProductSearchResult>)` — Results available.
  - `empty` — Search completed with no results.
  - `error(message: String)` — Search failed.

## State Transitions (Receipt UI)

```
Idle → Capturing → [capture] → Uploading(0..1)
                                  ├─ success → [poll receipt] → ReviewReady
                                  ├─ duplicate → [show existing]
                                  └─ error → [retry button] → Uploading

ReviewReady → [edit fields] → Confirming
                                 ├─ confirmed → PointsEarned
                                 └─ error → [show error, preserve edits]
```

## Validation Rules (Client-Side)

- **Currency required**: Every `ReceiptItem` must have non-null `currency` before confirmation (FR-011).
- **Email format**: Standard email regex validation on login/register.
- **Password minimum**: 8 characters on registration.
- **Image size**: Client-side check before upload (warn if >10MB).

## Serialization

All models use `json_serializable` with `@JsonSerializable()` annotation. Field names use `@JsonKey(name: 'snake_case_field')` to map JSON snake_case to Dart camelCase. Domain entities are separate from data models — data models are raw API shapes, entities are UI-optimized representations.
