# Data Model: Receipt OCR Processing and Review

## Receipt

- **Purpose**: Tracks uploaded receipt and lifecycle state.
- **Fields**:
  - `id`
  - `user_id`
  - `store_id` (nullable until resolved)
  - `image_url`
  - `image_hash` (duplicate detection)
  - `status` (`PENDING`, `NEEDS_REVIEW`, `CONFIRMED`, `REJECTED`)
  - `purchase_date` (nullable)
  - `total` (nullable)
  - `created_at`, `updated_at`

## ReceiptItem

- **Purpose**: Parsed or corrected line items linked to a receipt.
- **Fields**:
  - `id`
  - `receipt_id`
  - `raw_text`
  - `normalized_name` (nullable)
  - `product_id` (nullable before normalization)
  - `quantity`
  - `unit_price`
  - `currency` (mandatory at confirmation)
  - `line_total`

## Store

- **Purpose**: Merchant reference for grouping observations.
- **Fields**:
  - `id`
  - `name`
  - `branch` (nullable)
  - `address` (nullable)
  - `latitude` (nullable)
  - `longitude` (nullable)

## Product

- **Purpose**: Canonical product identity for normalization.
- **Fields**:
  - `id`
  - `canonical_name`
  - `unit` (nullable)

## PriceObservation

- **Purpose**: Confirmed observation for analytics/ranking.
- **Fields**:
  - `id`
  - `product_id`
  - `store_id`
  - `receipt_id`
  - `unit_price`
  - `currency` (NOT NULL, validated)
  - `observed_at`

## OCRJob

- **Purpose**: Queue processing tracker.
- **Fields**:
  - `id`
  - `receipt_id`
  - `status` (`QUEUED`, `PROCESSING`, `DONE`, `FAILED`)
  - `attempt`
  - `last_error`
  - `created_at`, `processed_at`

## RawOCRResult (Domain Value Object)

- **Purpose**: Raw extraction payload from `OCRProvider` before parser normalization.
- **Fields**:
  - `raw_text`
  - `lines[]`
  - `confidence` (optional)

## State Transitions

- `PENDING -> NEEDS_REVIEW` (after processing attempt)
- `NEEDS_REVIEW -> CONFIRMED` (successful user confirmation)
- `PENDING -> REJECTED` or `NEEDS_REVIEW -> REJECTED` (explicit rejection/failure policy)
- Other transitions are invalid.
