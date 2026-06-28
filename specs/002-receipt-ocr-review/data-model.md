# Data Model: Receipt OCR Review Flow

## Entities

### Receipt
- **Fields**: `id`, `user_id`, `store_id?`, `image_url`, `image_hash`, `status`, `purchase_date?`, `total?`, `created_at`, `updated_at`
- **Status lifecycle**: `PENDING -> NEEDS_REVIEW -> CONFIRMED` (or `REJECTED` if explicitly discarded)
- **Rules**:
  - Created in `PENDING` on upload.
  - Moves to `NEEDS_REVIEW` after OCR+parsing attempt (including unreadable OCR case).
  - Moves to `CONFIRMED` only after valid corrected payload.

### OCRJob
- **Fields**: `id`, `receipt_id`, `status`, `attempt_count`, `last_error?`, `created_at`, `processed_at?`
- **Rules**:
  - One active job per new receipt upload.
  - Duplicate same-user same-image uploads create no new job.

### Store
- **Fields**: `id`, `name`, `branch?`, `address?`, `lat?`, `lng?`, `created_at`
- **Rules**:
  - Merchant parser may propose existing store match.
  - Unknown merchant can create new store during review/confirmation.

### ReceiptItem
- **Fields**: `id`, `receipt_id`, `raw_text`, `normalized_name?`, `product_id?`, `quantity?`, `unit_price?`, `currency?`, `line_total?`
- **Rules**:
  - May be partial/empty in unreadable OCR scenario.
  - Becomes authoritative only after user confirmation.

### Product
- **Fields**: `id`, `normalized_name`, `category?`, `unit?`, `created_at`
- **Rules**:
  - Confirmation step maps each confirmed item to canonical product identity.

### PriceObservation
- **Fields**: `id`, `product_id`, `store_id`, `unit_price`, `currency`, `observed_at`, `receipt_id`
- **Rules**:
  - Created only on confirmed receipts.
  - Currency is mandatory; records without currency are invalid.

## Relationships

- `User 1..* Receipt`
- `Receipt 1..* ReceiptItem`
- `Receipt *..1 Store` (nullable before store extraction/review)
- `ReceiptItem *..1 Product` (nullable before confirmation)
- `Receipt 1..* PriceObservation` (at confirmation)
- `Store 1..* PriceObservation`
- `Product 1..* PriceObservation`

## Identity & Uniqueness

- Receipt duplicate key candidate: `(user_id, image_hash)` for idempotent duplicate detection.
- Price observation uniqueness heuristic: `(receipt_id, product_id, currency, unit_price, store_id)` to prevent accidental duplicate insertions on retry.

## Validation Rules

- Upload requires authenticated user and valid image payload.
- Confirmation requires editable payload consistency and mandatory currency per item intended for observation creation.
- Confirmation fails atomically if any required item currency is missing.
